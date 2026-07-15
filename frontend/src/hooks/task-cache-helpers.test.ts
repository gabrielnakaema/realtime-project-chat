import { QueryClient } from '@tanstack/react-query';
import { describe, expect, it } from 'vitest';
import {
  adjustBoardColumnCounts,
  buildColumnQueryKey,
  findTaskInBoardCaches,
  reconcileTaskInBoard,
} from './task-cache-helpers';
import type { Task } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';

const makeTask = (id: string, columnId: string, overrides: Partial<Task> = {}): Task =>
  ({
    id,
    project_id: 'project-1',
    project_column_id: columnId,
    title: `Task ${id}`,
    order: 'a0',
    version: 0,
    archived_at: null,
    ...overrides,
  }) as Task;

const seedColumn = (queryClient: QueryClient, projectId: string, columnId: string, tasks: Task[]) => {
  queryClient.setQueryData(buildColumnQueryKey(projectId, columnId), {
    pageParams: [{ taskOrder: '', updatedAt: null }],
    pages: [
      {
        [columnId]: { data: tasks, has_next: false },
      },
    ],
  });
};

const columnTaskIds = (queryClient: QueryClient, projectId: string, columnId: string): string[] => {
  const data = queryClient.getQueryData<{ pages: Record<string, { data: Task[] }>[] }>(
    buildColumnQueryKey(projectId, columnId),
  );
  return (data?.pages ?? []).flatMap((page) => page[columnId].data.map((t) => t.id));
};

const readCounts = (queryClient: QueryClient, projectId: string, columnIds: string[]) =>
  queryClient.getQueryData<Record<string, number>>(taskQueryKeys.countByColumn(projectId, columnIds));

describe('findTaskInBoardCaches', () => {
  it('finds a task in a column the caller did not name (the source column)', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [makeTask('task-1', 'column-pending')]);
    seedColumn(queryClient, 'project-1', 'column-doing', [makeTask('task-2', 'column-doing')]);

    const found = findTaskInBoardCaches(queryClient, 'project-1', 'task-1');

    expect(found).toBeDefined();
    expect(found?.columnId).toBe('column-pending');
    expect(found?.task.id).toBe('task-1');
  });

  it('returns undefined when the task is not on the board', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [makeTask('task-1', 'column-pending')]);

    expect(findTaskInBoardCaches(queryClient, 'project-1', 'missing')).toBeUndefined();
  });

  it('does not match tasks from a different project', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-2', 'column-pending', [makeTask('task-1', 'column-pending')]);

    expect(findTaskInBoardCaches(queryClient, 'project-1', 'task-1')).toBeUndefined();
  });
});

describe('adjustBoardColumnCounts', () => {
  it('applies deltas to the project count cache without knowing its column-id key', () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(taskQueryKeys.countByColumn('project-1', ['column-pending', 'column-doing']), {
      'column-pending': 3,
      'column-doing': 1,
    });

    adjustBoardColumnCounts(queryClient, 'project-1', { 'column-pending': -1, 'column-doing': 1 });

    const counts = queryClient.getQueryData<Record<string, number>>(
      taskQueryKeys.countByColumn('project-1', ['column-pending', 'column-doing']),
    );
    expect(counts).toEqual({ 'column-pending': 2, 'column-doing': 2 });
  });

  it('never drives a count below zero', () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(taskQueryKeys.countByColumn('project-1', ['column-pending']), { 'column-pending': 0 });

    adjustBoardColumnCounts(queryClient, 'project-1', { 'column-pending': -1 });

    expect(
      queryClient.getQueryData<Record<string, number>>(taskQueryKeys.countByColumn('project-1', ['column-pending'])),
    ).toEqual({ 'column-pending': 0 });
  });

  it('leaves a different project untouched', () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(taskQueryKeys.countByColumn('project-2', ['column-pending']), { 'column-pending': 5 });

    adjustBoardColumnCounts(queryClient, 'project-1', { 'column-pending': -1 });

    expect(
      queryClient.getQueryData<Record<string, number>>(taskQueryKeys.countByColumn('project-2', ['column-pending'])),
    ).toEqual({ 'column-pending': 5 });
  });
});

describe('reconcileTaskInBoard', () => {
  const columns = ['column-pending', 'column-doing'];
  const seedBoard = (queryClient: QueryClient, pending: Task[], doing: Task[], counts: Record<string, number>) => {
    seedColumn(queryClient, 'project-1', 'column-pending', pending);
    seedColumn(queryClient, 'project-1', 'column-doing', doing);
    queryClient.setQueryData(taskQueryKeys.countByColumn('project-1', columns), counts);
  };

  it('inserts a task new to the board and increments its column count', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [], [], { 'column-pending': 0, 'column-doing': 0 });

    reconcileTaskInBoard(queryClient, 'project-1', makeTask('task-1', 'column-pending', { version: 1 }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1']);
    expect(readCounts(queryClient, 'project-1', columns)).toEqual({ 'column-pending': 1, 'column-doing': 0 });
  });

  it('repositions a same-column update without touching counts', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { order: 'a0', version: 1 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTaskInBoard(
      queryClient,
      'project-1',
      makeTask('task-1', 'column-pending', { order: 'a5', version: 2, title: 'renamed' }),
    );

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1']);
    expect(readCounts(queryClient, 'project-1', columns)).toEqual({ 'column-pending': 1, 'column-doing': 0 });
  });

  it('moves a task across columns and shifts both counts by one', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 1 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTaskInBoard(queryClient, 'project-1', makeTask('task-1', 'column-doing', { version: 2 }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual(['task-1']);
    expect(readCounts(queryClient, 'project-1', columns)).toEqual({ 'column-pending': 0, 'column-doing': 1 });
  });

  it('removes an archived task from its column and decrements the count', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 1 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTaskInBoard(
      queryClient,
      'project-1',
      makeTask('task-1', 'column-pending', { version: 2, archived_at: '2026-07-15T00:00:00Z' }),
    );

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(readCounts(queryClient, 'project-1', columns)).toEqual({ 'column-pending': 0, 'column-doing': 0 });
  });

  it('drops a stale/echoed update whose version is not newer', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 5, title: 'current' })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTaskInBoard(queryClient, 'project-1', makeTask('task-1', 'column-doing', { version: 5 }));
    reconcileTaskInBoard(queryClient, 'project-1', makeTask('task-1', 'column-doing', { version: 4 }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1']);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual([]);
    expect(readCounts(queryClient, 'project-1', columns)).toEqual({ 'column-pending': 1, 'column-doing': 0 });
  });

  it('force applies a same-version write (the optimistic move path)', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 3 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTaskInBoard(queryClient, 'project-1', makeTask('task-1', 'column-doing', { version: 3 }), { force: true });

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual(['task-1']);
    expect(readCounts(queryClient, 'project-1', columns)).toEqual({ 'column-pending': 0, 'column-doing': 1 });
  });

  it('uses previousColumnId to fix counts when the source page is not loaded', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [], [], { 'column-pending': 1, 'column-doing': 0 });

    reconcileTaskInBoard(queryClient, 'project-1', makeTask('task-1', 'column-doing', { version: 2 }), {
      previousColumnId: 'column-pending',
    });

    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual(['task-1']);
    expect(readCounts(queryClient, 'project-1', columns)).toEqual({ 'column-pending': 0, 'column-doing': 1 });
  });
});

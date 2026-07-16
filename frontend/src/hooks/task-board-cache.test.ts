import { QueryClient } from '@tanstack/react-query';
import { describe, expect, it } from 'vitest';
import { compareTasks, findTaskOnBoard, reconcileTask, sortedColumnTasks, upsertTask } from './task-board-cache';
import type { BoardColumnPage, ColumnCache } from './task-board-cache';
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
    updated_at: '2026-07-01T00:00:00Z',
    ...overrides,
  }) as Task;

const makePage = (tasks: Task[], hasNext = false): BoardColumnPage => ({
  data: tasks,
  has_next: hasNext,
  has_previous: false,
  cursor: null,
});

const seedColumn = (
  queryClient: QueryClient,
  projectId: string,
  columnId: string,
  tasks: Task[],
  options: { hasNext?: boolean } = {},
) => {
  queryClient.setQueryData<ColumnCache>(taskQueryKeys.boardColumn(projectId, columnId), {
    pageParams: [{ taskOrder: '', updatedAt: null }],
    pages: [makePage(tasks, options.hasNext ?? false)],
  });
};

const seedCounts = (queryClient: QueryClient, projectId: string, columnIds: string[], counts: Record<string, number>) =>
  queryClient.setQueryData(taskQueryKeys.countsFor(projectId, columnIds), counts);

const columnTaskIds = (queryClient: QueryClient, projectId: string, columnId: string): string[] =>
  sortedColumnTasks(queryClient.getQueryData<ColumnCache>(taskQueryKeys.boardColumn(projectId, columnId))).map(
    (t) => t.id,
  );

const countsInvalidated = (queryClient: QueryClient, projectId: string, columnIds: string[]) =>
  queryClient.getQueryState(taskQueryKeys.countsFor(projectId, columnIds))?.isInvalidated ?? false;

const seedArchived = (queryClient: QueryClient, projectId: string) =>
  queryClient.setQueryData(taskQueryKeys.archived(projectId), { pageParams: [], pages: [] });

const archivedInvalidated = (queryClient: QueryClient, projectId: string) =>
  queryClient.getQueryState(taskQueryKeys.archived(projectId))?.isInvalidated ?? false;

describe('compareTasks / sortedColumnTasks', () => {
  it('sorts by order asc, then updated_at desc, then id (mirrors the backend ORDER BY)', () => {
    const cache: Pick<ColumnCache, 'pages'> = {
      pages: [
        makePage([
          makeTask('task-c', 'col', { order: 'b0' }),
          makeTask('task-b', 'col', { order: 'a0', updated_at: '2026-07-01T00:00:00Z' }),
        ]),
        makePage([
          makeTask('task-a', 'col', { order: 'a0', updated_at: '2026-07-02T00:00:00Z' }),
          makeTask('task-d', 'col', { order: 'a0', updated_at: '2026-07-01T00:00:00Z' }),
        ]),
      ],
    };

    expect(sortedColumnTasks(cache).map((t) => t.id)).toEqual(['task-a', 'task-b', 'task-d', 'task-c']);
  });

  it('returns an empty list for a missing cache', () => {
    expect(sortedColumnTasks(undefined)).toEqual([]);
  });

  it('compareTasks is a strict weak ordering on identical tasks', () => {
    const task = makeTask('task-1', 'col');
    expect(compareTasks(task, task)).toBe(0);
  });
});

describe('findTaskOnBoard', () => {
  it('finds a task in a column the caller did not name (the source column)', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [makeTask('task-1', 'column-pending')]);
    seedColumn(queryClient, 'project-1', 'column-doing', [makeTask('task-2', 'column-doing')]);

    const found = findTaskOnBoard(queryClient, 'project-1', 'task-1');

    expect(found).toBeDefined();
    expect(found?.columnId).toBe('column-pending');
    expect(found?.task.id).toBe('task-1');
  });

  it('returns undefined when the task is not on the board', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [makeTask('task-1', 'column-pending')]);

    expect(findTaskOnBoard(queryClient, 'project-1', 'missing')).toBeUndefined();
  });

  it('does not match tasks from a different project', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-2', 'column-pending', [makeTask('task-1', 'column-pending')]);

    expect(findTaskOnBoard(queryClient, 'project-1', 'task-1')).toBeUndefined();
  });
});

describe('upsertTask', () => {
  it('replaces an already-loaded task wherever it lives across pages', () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData<ColumnCache>(taskQueryKeys.boardColumn('project-1', 'column-pending'), {
      pageParams: [
        { taskOrder: '', updatedAt: null },
        { taskOrder: 'a1', updatedAt: null },
      ],
      pages: [
        makePage(
          [
            makeTask('task-1', 'column-pending', { order: 'a0' }),
            makeTask('task-2', 'column-pending', { order: 'a1' }),
          ],
          true,
        ),
        makePage([makeTask('task-3', 'column-pending', { order: 'a2' })], false),
      ],
    });

    upsertTask(
      queryClient,
      'project-1',
      'column-pending',
      makeTask('task-1', 'column-pending', { order: 'a1V', title: 'moved down' }),
    );

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-2', 'task-1', 'task-3']);
    const cache = queryClient.getQueryData<ColumnCache>(taskQueryKeys.boardColumn('project-1', 'column-pending'));
    expect(cache?.pages[0].data.map((t) => t.id)).toEqual(['task-2']);
    expect(cache?.pages[1].data.map((t) => t.id)).toEqual(['task-3', 'task-1']);
  });

  it('drops a task whose order is beyond the loaded window of a paginated column', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [makeTask('task-1', 'column-pending', { order: 'a0' })], {
      hasNext: true,
    });

    upsertTask(queryClient, 'project-1', 'column-pending', makeTask('task-new', 'column-pending', { order: 'z0' }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1']);
  });

  it('removes a loaded task that moved beyond the loaded window', () => {
    const queryClient = new QueryClient();
    seedColumn(
      queryClient,
      'project-1',
      'column-pending',
      [makeTask('task-1', 'column-pending', { order: 'a0' }), makeTask('task-2', 'column-pending', { order: 'a1' })],
      { hasNext: true },
    );

    upsertTask(queryClient, 'project-1', 'column-pending', makeTask('task-1', 'column-pending', { order: 'z0' }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-2']);
  });

  it('inserts beyond the loaded window when allowBeyondWindow is set (optimistic move)', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [makeTask('task-1', 'column-pending', { order: 'a0' })], {
      hasNext: true,
    });

    upsertTask(queryClient, 'project-1', 'column-pending', makeTask('task-new', 'column-pending', { order: 'z0' }), {
      allowBeyondWindow: true,
    });

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1', 'task-new']);
  });

  it('inserts normally into a fully-loaded column', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [makeTask('task-1', 'column-pending', { order: 'a0' })]);

    upsertTask(queryClient, 'project-1', 'column-pending', makeTask('task-new', 'column-pending', { order: 'z0' }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1', 'task-new']);
  });
});

describe('reconcileTask', () => {
  const columns = ['column-pending', 'column-doing'];
  const seedBoard = (queryClient: QueryClient, pending: Task[], doing: Task[], counts: Record<string, number>) => {
    seedColumn(queryClient, 'project-1', 'column-pending', pending);
    seedColumn(queryClient, 'project-1', 'column-doing', doing);
    seedCounts(queryClient, 'project-1', columns, counts);
    seedArchived(queryClient, 'project-1');
  };

  it('inserts a task new to the board and invalidates counts and the archived list (possible restore)', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [], [], { 'column-pending': 0, 'column-doing': 0 });

    reconcileTask(queryClient, makeTask('task-1', 'column-pending', { version: 1 }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1']);
    expect(countsInvalidated(queryClient, 'project-1', columns)).toBe(true);
    expect(archivedInvalidated(queryClient, 'project-1')).toBe(true);
  });

  it('applies a same-column update without invalidating counts', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { order: 'a0', version: 1 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTask(queryClient, makeTask('task-1', 'column-pending', { order: 'a5', version: 2, title: 'renamed' }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1']);
    expect(findTaskOnBoard(queryClient, 'project-1', 'task-1')?.task.title).toBe('renamed');
    expect(countsInvalidated(queryClient, 'project-1', columns)).toBe(false);
    expect(archivedInvalidated(queryClient, 'project-1')).toBe(false);
  });

  it('moves a task across columns and invalidates counts', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 1 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTask(queryClient, makeTask('task-1', 'column-doing', { version: 2 }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual(['task-1']);
    expect(countsInvalidated(queryClient, 'project-1', columns)).toBe(true);
    expect(archivedInvalidated(queryClient, 'project-1')).toBe(false);
  });

  it('removes an archived task from its column and invalidates counts and the archived list', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 1 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTask(
      queryClient,
      makeTask('task-1', 'column-pending', { version: 2, archived_at: '2026-07-15T00:00:00Z' }),
    );

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(countsInvalidated(queryClient, 'project-1', columns)).toBe(true);
    expect(archivedInvalidated(queryClient, 'project-1')).toBe(true);
  });

  it('drops a stale/echoed update whose version is not newer', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 5, title: 'current' })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTask(queryClient, makeTask('task-1', 'column-doing', { version: 5 }));
    reconcileTask(queryClient, makeTask('task-1', 'column-doing', { version: 4 }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-1']);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual([]);
    expect(countsInvalidated(queryClient, 'project-1', columns)).toBe(false);
  });

  it('does not resurrect an archived task when a delayed lower-version event arrives after removal', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 4 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTask(
      queryClient,
      makeTask('task-1', 'column-pending', { version: 6, archived_at: '2026-07-16T00:00:00Z' }),
    );
    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);

    reconcileTask(queryClient, makeTask('task-1', 'column-pending', { version: 5, title: 'stale edit' }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
  });

  it('does not reinsert stale state after a task moved beyond the loaded window', () => {
    const queryClient = new QueryClient();
    seedColumn(
      queryClient,
      'project-1',
      'column-pending',
      [
        makeTask('task-1', 'column-pending', { order: 'a0', version: 2 }),
        makeTask('task-2', 'column-pending', { order: 'a1' }),
      ],
      { hasNext: true },
    );

    reconcileTask(queryClient, makeTask('task-1', 'column-pending', { order: 'z0', version: 3 }));
    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-2']);

    reconcileTask(queryClient, makeTask('task-1', 'column-pending', { order: 'a05', version: 2 }));

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual(['task-2']);
  });

  it('optimistic write bypasses the version gate and skips count invalidation', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [makeTask('task-1', 'column-pending', { version: 3 })], [], {
      'column-pending': 1,
      'column-doing': 0,
    });

    reconcileTask(queryClient, makeTask('task-1', 'column-doing', { version: 3 }), { optimistic: true });

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual(['task-1']);
    expect(countsInvalidated(queryClient, 'project-1', columns)).toBe(false);
    expect(archivedInvalidated(queryClient, 'project-1')).toBe(false);
  });

  it('optimistic move stays visible even when the provisional order is beyond the loaded window', () => {
    const queryClient = new QueryClient();
    seedColumn(queryClient, 'project-1', 'column-pending', [
      makeTask('task-1', 'column-pending', { order: 'zzz', version: 3 }),
    ]);
    seedColumn(queryClient, 'project-1', 'column-doing', [makeTask('task-2', 'column-doing', { order: 'a0' })], {
      hasNext: true,
    });
    seedCounts(queryClient, 'project-1', columns, { 'column-pending': 1, 'column-doing': 1 });

    reconcileTask(queryClient, makeTask('task-1', 'column-doing', { order: 'zzz', version: 3 }), { optimistic: true });

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual(['task-2', 'task-1']);
  });

  it('archives via previousColumnId when the task is not cached', () => {
    const queryClient = new QueryClient();
    seedBoard(queryClient, [], [], { 'column-pending': 1, 'column-doing': 0 });

    reconcileTask(
      queryClient,
      makeTask('task-1', 'column-doing', { version: 2, archived_at: '2026-07-15T00:00:00Z' }),
      { previousColumnId: 'column-pending' },
    );

    expect(columnTaskIds(queryClient, 'project-1', 'column-pending')).toEqual([]);
    expect(columnTaskIds(queryClient, 'project-1', 'column-doing')).toEqual([]);
    expect(countsInvalidated(queryClient, 'project-1', columns)).toBe(true);
  });
});

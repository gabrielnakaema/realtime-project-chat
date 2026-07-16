// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook } from '@testing-library/react';
import { createElement } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useMoveTask } from './use-move-task';
import { findTaskOnBoard, reconcileTask, sortedColumnTasks } from './task-board-cache';
import type { ColumnCache } from './task-board-cache';
import type { Task } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';
import { moveTask } from '@/services/tasks';

vi.mock('@/services/tasks', () => ({
  moveTask: vi.fn(),
}));

const makeTask = (id: string, columnId: string, overrides: Partial<Task> = {}): Task =>
  ({
    id,
    project_id: 'project-1',
    project_column_id: columnId,
    title: `Task ${id}`,
    order: 'a0',
    version: 1,
    archived_at: null,
    updated_at: '2026-07-01T00:00:00Z',
    ...overrides,
  }) as Task;

const seedColumn = (queryClient: QueryClient, columnId: string, tasks: Task[], hasNext = false) => {
  queryClient.setQueryData<ColumnCache>(taskQueryKeys.boardColumn('project-1', columnId), {
    pageParams: [{ taskOrder: '', updatedAt: null }],
    pages: [{ data: tasks, has_next: hasNext, has_previous: false, cursor: null }],
  });
};

const columnTaskIds = (queryClient: QueryClient, columnId: string): string[] =>
  sortedColumnTasks(queryClient.getQueryData<ColumnCache>(taskQueryKeys.boardColumn('project-1', columnId))).map(
    (t) => t.id,
  );

const renderMoveTask = (queryClient: QueryClient) => {
  const wrapper = ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
  return renderHook(() => useMoveTask(), { wrapper });
};

afterEach(() => vi.clearAllMocks());

describe('useMoveTask', () => {
  it('onError invalidates the project board and counts instead of force-writing the stale snapshot', async () => {
    const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    seedColumn(queryClient, 'col-a', [makeTask('task-1', 'col-a')]);
    seedColumn(queryClient, 'col-b', []);
    queryClient.setQueryData(taskQueryKeys.countsFor('project-1', ['col-a', 'col-b']), { 'col-a': 1, 'col-b': 0 });

    vi.mocked(moveTask).mockRejectedValue(new Error('network down'));

    const { result } = renderMoveTask(queryClient);
    await result.current
      .mutateAsync({ taskId: 'task-1', projectId: 'project-1', projectColumnId: 'col-b', afterTaskId: null })
      .catch(() => {});

    expect(queryClient.getQueryState(taskQueryKeys.boardColumn('project-1', 'col-a'))?.isInvalidated).toBe(true);
    expect(queryClient.getQueryState(taskQueryKeys.boardColumn('project-1', 'col-b'))?.isInvalidated).toBe(true);
    expect(queryClient.getQueryState(taskQueryKeys.countsFor('project-1', ['col-a', 'col-b']))?.isInvalidated).toBe(
      true,
    );
  });

  it('onMutate places the task between its drop neighbors via a client-computed fractional order', async () => {
    const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    seedColumn(queryClient, 'col-a', [makeTask('task-moving', 'col-a', { order: 'm0' })]);
    seedColumn(queryClient, 'col-b', [
      makeTask('task-top', 'col-b', { order: 'a0' }),
      makeTask('task-bottom', 'col-b', { order: 'c0' }),
    ]);

    let resolveMove: (task: Task) => void = () => {};
    vi.mocked(moveTask).mockImplementation(() => new Promise<Task>((resolve) => (resolveMove = resolve)));

    const { result } = renderMoveTask(queryClient);
    const pending = result.current.mutateAsync({
      taskId: 'task-moving',
      projectId: 'project-1',
      projectColumnId: 'col-b',
      afterTaskId: 'task-top',
    });

    await vi.waitFor(() => {
      expect(columnTaskIds(queryClient, 'col-b')).toEqual(['task-top', 'task-moving', 'task-bottom']);
    });
    expect(columnTaskIds(queryClient, 'col-a')).toEqual([]);
    // generateKeyBetween('a0', 'c0') === 'b' — same math the server runs.
    expect(findTaskOnBoard(queryClient, 'project-1', 'task-moving')?.task.order).toBe('b');

    resolveMove(makeTask('task-moving', 'col-b', { order: 'b', version: 2 }));
    await pending;
    expect(columnTaskIds(queryClient, 'col-b')).toEqual(['task-top', 'task-moving', 'task-bottom']);
  });

  it('invalidates counts after a successful cross-column move despite the optimistic pre-placement', async () => {
    const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    seedColumn(queryClient, 'col-a', [makeTask('task-1', 'col-a', { version: 1 })]);
    seedColumn(queryClient, 'col-b', []);
    queryClient.setQueryData(taskQueryKeys.countsFor('project-1', ['col-a', 'col-b']), { 'col-a': 1, 'col-b': 0 });

    vi.mocked(moveTask).mockResolvedValue(makeTask('task-1', 'col-b', { order: 'b0', version: 2 }));

    const { result } = renderMoveTask(queryClient);
    await result.current.mutateAsync({
      taskId: 'task-1',
      projectId: 'project-1',
      projectColumnId: 'col-b',
      afterTaskId: null,
    });

    expect(queryClient.getQueryState(taskQueryKeys.countsFor('project-1', ['col-a', 'col-b']))?.isInvalidated).toBe(
      true,
    );
  });

  it('does not invalidate counts on a same-column reorder', async () => {
    const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    seedColumn(queryClient, 'col-a', [
      makeTask('task-1', 'col-a', { order: 'a0', version: 1 }),
      makeTask('task-2', 'col-a', { order: 'c0', version: 1 }),
    ]);
    queryClient.setQueryData(taskQueryKeys.countsFor('project-1', ['col-a']), { 'col-a': 2 });

    vi.mocked(moveTask).mockResolvedValue(makeTask('task-1', 'col-a', { order: 'd0', version: 2 }));

    const { result } = renderMoveTask(queryClient);
    await result.current.mutateAsync({
      taskId: 'task-1',
      projectId: 'project-1',
      projectColumnId: 'col-a',
      afterTaskId: 'task-2',
    });

    expect(queryClient.getQueryState(taskQueryKeys.countsFor('project-1', ['col-a']))?.isInvalidated).toBe(false);
  });

  it('suppresses the websocket echo of its own move via the version gate', async () => {
    const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    seedColumn(queryClient, 'col-a', [makeTask('task-1', 'col-a', { version: 3 })]);
    seedColumn(queryClient, 'col-b', []);

    vi.mocked(moveTask).mockResolvedValue(makeTask('task-1', 'col-b', { order: 'b0', version: 4 }));

    const { result } = renderMoveTask(queryClient);
    await result.current.mutateAsync({
      taskId: 'task-1',
      projectId: 'project-1',
      projectColumnId: 'col-b',
      afterTaskId: null,
    });

    expect(findTaskOnBoard(queryClient, 'project-1', 'task-1')?.task.version).toBe(4);

    // Simulate the websocket echo: same version, would move it back if applied.
    reconcileTask(queryClient, makeTask('task-1', 'col-a', { order: 'z0', version: 4 }), {
      previousColumnId: 'col-a',
    });

    expect(columnTaskIds(queryClient, 'col-b')).toEqual(['task-1']);
    expect(columnTaskIds(queryClient, 'col-a')).toEqual([]);
  });

  it('keeps a card dropped at the end of a partially-loaded column visible until the server responds', async () => {
    const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    seedColumn(queryClient, 'col-a', [makeTask('task-moving', 'col-a', { order: 'm0' })]);
    seedColumn(queryClient, 'col-b', [makeTask('task-loaded', 'col-b', { order: 'a0' })], true);

    vi.mocked(moveTask).mockImplementation(() => new Promise<Task>(() => {}));

    const { result } = renderMoveTask(queryClient);
    result.current
      .mutateAsync({
        taskId: 'task-moving',
        projectId: 'project-1',
        projectColumnId: 'col-b',
        afterTaskId: 'task-loaded',
      })
      .catch(() => {});

    await vi.waitFor(() => {
      expect(columnTaskIds(queryClient, 'col-b')).toEqual(['task-loaded', 'task-moving']);
    });
  });
});

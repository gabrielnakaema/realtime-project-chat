// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { createElement } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useBoardColumnTasks } from './use-board-column-tasks';
import { upsertTask } from './task-board-cache';
import type { CursorPaginated } from '@/types/paginated';
import type { Task } from '@/types/task';
import { listColumnTasks } from '@/services/tasks';

vi.mock('@/services/tasks', () => ({
  listColumnTasks: vi.fn(),
}));

vi.mock('@/hooks/use-infinite-scroll-observer', () => ({
  useInfiniteScrollObserver: () => () => {},
}));

const makeTask = (id: string, order: string): Task =>
  ({
    id,
    project_id: 'project-1',
    project_column_id: 'col-a',
    title: `Task ${id}`,
    order,
    version: 1,
    archived_at: null,
    updated_at: '2026-07-01T00:00:00Z',
  }) as Task;

const makeResponse = (tasks: Task[], hasNext: boolean): CursorPaginated<Task> => ({
  data: tasks,
  has_next: hasNext,
  has_previous: false,
});

const renderBoardColumn = (queryClient: QueryClient) => {
  const wrapper = ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
  return renderHook(() => useBoardColumnTasks({ projectId: 'project-1', columnId: 'col-a', limit: 2 }), { wrapper });
};

afterEach(() => vi.clearAllMocks());

describe('useBoardColumnTasks pagination cursor', () => {
  it('derives the next-page cursor from the server response, not from optimistic cache contents', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    vi.mocked(listColumnTasks)
      .mockResolvedValueOnce(makeResponse([makeTask('task-1', 'a0'), makeTask('task-2', 'a1')], true))
      .mockResolvedValueOnce(makeResponse([makeTask('task-3', 'a2')], false));

    const { result } = renderBoardColumn(queryClient);
    await waitFor(() => expect(result.current.columnTasks.map((t) => t.id)).toEqual(['task-1', 'task-2']));

    upsertTask(queryClient, 'project-1', 'col-a', makeTask('task-optimistic', 'z9'), { allowBeyondWindow: true });
    await waitFor(() =>
      expect(result.current.columnTasks.map((t) => t.id)).toEqual(['task-1', 'task-2', 'task-optimistic']),
    );

    await result.current.queryResult.fetchNextPage();

    expect(vi.mocked(listColumnTasks).mock.calls[1][0]).toMatchObject({ taskOrder: 'a1' });
    await waitFor(() =>
      expect(result.current.columnTasks.map((t) => t.id)).toEqual(['task-1', 'task-2', 'task-3', 'task-optimistic']),
    );
  });

  it('stops paginating when the server reports no next page', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    vi.mocked(listColumnTasks).mockResolvedValueOnce(makeResponse([makeTask('task-1', 'a0')], false));

    const { result } = renderBoardColumn(queryClient);
    await waitFor(() => expect(result.current.columnTasks.map((t) => t.id)).toEqual(['task-1']));

    expect(result.current.queryResult.hasNextPage).toBe(false);
  });
});

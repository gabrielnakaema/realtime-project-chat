// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useRealtimeTaskSync } from './use-realtime-task-sync';
import type { ReactNode } from 'react';
import type { Task } from '@/features/tasks/types/task';
import type { SocketEvent } from '@/shared/types/websocket';
import { useSocket } from '@/shared/hooks/use-socket';
import { taskQueryKeys } from '@/shared/services/query-keys';

vi.mock('@/shared/hooks/use-socket', () => ({
  useSocket: vi.fn(),
}));

const mockUseSocket = vi.mocked(useSocket);

const makeTask = (overrides: Partial<Task> = {}): Task =>
  ({
    id: 'task-1',
    project_id: 'project-1',
    project_column_id: 'column-1',
    title: 'Current title',
    order: 'a0',
    version: 1,
    archived_at: null,
    updated_at: '2026-07-25T12:00:00Z',
    updates: [],
    ...overrides,
  }) as Task;

afterEach(() => {
  vi.clearAllMocks();
});

describe('useRealtimeTaskSync', () => {
  it('reconciles and refreshes task details when a task_updated event arrives', async () => {
    let handleEvent: ((event: SocketEvent) => void) | undefined;
    const subscribe = vi.fn((_roomId: string, _type: string, handler: (event: SocketEvent) => void) => {
      handleEvent = handler;
      return vi.fn();
    });
    mockUseSocket.mockReturnValue({ status: 'connected', subscribe } as ReturnType<typeof useSocket>);

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const currentTask = makeTask();
    queryClient.setQueryData(taskQueryKeys.details(currentTask.id), currentTask);
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    renderHook(() => useRealtimeTaskSync('project-1'), { wrapper });
    await waitFor(() => expect(subscribe).toHaveBeenCalledWith('project-1', 'project', expect.any(Function)));

    act(() => {
      handleEvent?.({
        type: 'task_updated',
        room_id: 'project-1',
        data: {
          task: makeTask({ version: 2, title: 'Updated by collaborator' }),
        },
      });
    });

    expect(queryClient.getQueryData<Task>(taskQueryKeys.details(currentTask.id))?.title).toBe(
      'Updated by collaborator',
    );
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.details(currentTask.id),
      exact: true,
    });
  });
});

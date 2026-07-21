// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, waitFor } from '@testing-library/react';
import { toast } from 'react-toastify';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ProjectSync } from './project-sync';
import type { SocketEvent } from '@/shared/types/websocket';
import { useSocket } from '@/shared/hooks/use-socket';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { projectQueryKeys } from '@/shared/services/query-keys';

const mockNavigate = vi.fn();
let currentPathname = '/projects';

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
  useRouterState: (options: { select: (state: { location: { pathname: string } }) => string }) =>
    options.select({ location: { pathname: currentPathname } }),
}));

vi.mock('@/features/auth/hooks/use-auth', () => ({
  useAuth: vi.fn(),
}));

vi.mock('@/shared/hooks/use-socket', () => ({
  useSocket: vi.fn(),
}));

vi.mock('react-toastify', () => ({
  toast: {
    info: vi.fn(),
  },
}));

const mockUseAuth = vi.mocked(useAuth);
const mockUseSocket = vi.mocked(useSocket);
const mockToastInfo = vi.mocked(toast.info);

interface Subscription {
  roomId: string;
  type: string;
  handler: (event: SocketEvent) => void;
}

const renderProjectSync = () => {
  const subscriptions: Subscription[] = [];
  const unsubscribe = vi.fn();
  const subscribe = vi.fn((roomId: string, type: string, handler: (event: SocketEvent) => void) => {
    subscriptions.push({ roomId, type, handler });
    return unsubscribe;
  });

  mockUseAuth.mockReturnValue({ user: { id: 'user-1', email: 'user@example.com', name: 'User' } } as ReturnType<
    typeof useAuth
  >);
  mockUseSocket.mockReturnValue({ status: 'connected', subscribe } as unknown as ReturnType<typeof useSocket>);

  const queryClient = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
  const removeQueries = vi.spyOn(queryClient, 'removeQueries');

  render(
    <QueryClientProvider client={queryClient}>
      <ProjectSync />
    </QueryClientProvider>,
  );

  const dispatch = (roomId: string, type: string, event: SocketEvent) => {
    for (const subscription of subscriptions) {
      if (subscription.roomId === roomId && subscription.type === type) {
        subscription.handler(event);
      }
    }
  };

  return { subscriptions, subscribe, invalidateQueries, removeQueries, dispatch };
};

beforeEach(() => {
  currentPathname = '/projects';
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('ProjectSync', () => {
  it('invalidates queries and redirects when the current user is removed from the viewed project', () => {
    currentPathname = '/projects/project-1/settings/members';
    const { invalidateQueries, removeQueries, dispatch } = renderProjectSync();

    dispatch('user-1', 'user', {
      type: 'project_member_removed',
      room_id: 'user-1',
      data: { id: 'member-1', user_id: 'user-1', project_id: 'project-1', role: 'member', user: {} },
    } as SocketEvent);

    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['projects'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['tasks'] });
    expect(removeQueries).toHaveBeenCalledWith({ queryKey: projectQueryKeys.details('project-1') });
    expect(mockNavigate).toHaveBeenCalledWith({ to: '/projects' });
  });

  it('does not open a project room subscription when not viewing a project page', () => {
    currentPathname = '/dashboard';
    const { subscribe } = renderProjectSync();

    expect(subscribe).not.toHaveBeenCalledWith(expect.anything(), 'project', expect.anything());
  });

  it('subscribes to the current project room and reacts to project_deleted', async () => {
    currentPathname = '/projects/project-1/settings';
    const { subscribe, invalidateQueries, removeQueries, dispatch } = renderProjectSync();

    expect(subscribe).toHaveBeenCalledWith('project-1', 'project', expect.any(Function));

    dispatch('project-1', 'project', {
      type: 'project_deleted',
      room_id: 'project-1',
      data: { id: 'project-1', name: 'Deleted project' },
    } as unknown as SocketEvent);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/projects' });
    });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['projects'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['tasks'] });
    expect(removeQueries).toHaveBeenCalledWith({ queryKey: projectQueryKeys.details('project-1') });
    expect(removeQueries).toHaveBeenCalledWith({ queryKey: ['tasks'] });
    expect(mockToastInfo).toHaveBeenCalledWith('This project was deleted.');
  });
});

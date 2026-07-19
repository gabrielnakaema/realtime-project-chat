// @vitest-environment jsdom

import { act, cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SocketProvider } from './socket-context';
import type { ReactNode } from 'react';
import { useAuth } from '@/features/auth/hooks/use-auth';

vi.mock('@/features/auth/hooks/use-auth', () => ({
  useAuth: vi.fn(),
}));

const mockUseAuth = vi.mocked(useAuth);

class MockWebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
  static instances: MockWebSocket[] = [];

  readonly url: string;
  readyState = MockWebSocket.CONNECTING;
  onopen: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  send = vi.fn();

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  close() {
    this.readyState = MockWebSocket.CLOSED;
  }

  open() {
    this.readyState = MockWebSocket.OPEN;
    this.onopen?.(new Event('open'));
  }

  disconnect() {
    this.readyState = MockWebSocket.CLOSED;
    this.onclose?.(new CloseEvent('close'));
  }
}

const authenticated = {
  authStatus: 'authenticated' as const,
  authenticate: vi.fn(),
  isAuthenticated: true,
  logout: vi.fn(),
  user: undefined,
};

const unauthenticated = {
  ...authenticated,
  authStatus: 'unauthenticated' as const,
  isAuthenticated: false,
};

const renderProvider = (children: ReactNode = <div>Application</div>) =>
  render(<SocketProvider>{children}</SocketProvider>);

beforeEach(() => {
  vi.useFakeTimers();
  MockWebSocket.instances = [];
  vi.stubGlobal('WebSocket', MockWebSocket);
  mockUseAuth.mockReturnValue(authenticated);
});

afterEach(() => {
  cleanup();
  vi.runOnlyPendingTimers();
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});

describe('SocketProvider connection indicator', () => {
  it('stays hidden during the initial connection and while it is healthy', () => {
    renderProvider();

    expect(screen.queryByRole('status')).toBeNull();

    act(() => MockWebSocket.instances[0].open());

    expect(screen.queryByRole('status')).toBeNull();
  });

  it('shows the real retry delay and counts down across repeated attempts', () => {
    renderProvider();
    act(() => MockWebSocket.instances[0].open());

    act(() => MockWebSocket.instances[0].disconnect());

    expect(screen.getByRole('status').textContent).toContain('Reconnecting in 1 second');

    act(() => vi.advanceTimersByTime(1_000));
    expect(MockWebSocket.instances).toHaveLength(2);

    act(() => MockWebSocket.instances[1].disconnect());
    expect(screen.getByRole('status').textContent).toContain('Reconnecting in 2 seconds');

    act(() => vi.advanceTimersByTime(1_000));
    expect(screen.getByRole('status').textContent).toContain('Reconnecting in 1 second');

    act(() => vi.advanceTimersByTime(1_000));
    expect(MockWebSocket.instances).toHaveLength(3);
  });

  it('shows a successful reconnection for three seconds', () => {
    renderProvider();
    act(() => MockWebSocket.instances[0].open());
    act(() => MockWebSocket.instances[0].disconnect());
    act(() => vi.advanceTimersByTime(1_000));

    act(() => MockWebSocket.instances[1].open());

    expect(screen.getByRole('status').textContent).toContain('Connection restored');

    act(() => vi.advanceTimersByTime(2_999));
    expect(screen.getByRole('status').textContent).toContain('Connection restored');

    act(() => vi.advanceTimersByTime(1));
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('cancels pending reconnect work on logout and unmount', () => {
    const view = renderProvider();
    act(() => MockWebSocket.instances[0].open());
    act(() => MockWebSocket.instances[0].disconnect());

    mockUseAuth.mockReturnValue(unauthenticated);
    view.rerender(
      <SocketProvider>
        <div>Application</div>
      </SocketProvider>,
    );

    expect(screen.queryByRole('status')).toBeNull();
    act(() => vi.advanceTimersByTime(30_000));
    expect(MockWebSocket.instances).toHaveLength(1);

    mockUseAuth.mockReturnValue(authenticated);
    const mountedAgain = renderProvider();
    act(() => MockWebSocket.instances[1].disconnect());
    mountedAgain.unmount();
    act(() => vi.advanceTimersByTime(30_000));
    expect(MockWebSocket.instances).toHaveLength(2);
  });
});

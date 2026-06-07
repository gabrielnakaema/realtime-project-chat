// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { HeaderUser } from './header-user';
import type { ComponentProps, ReactNode } from 'react';
import { useAuth } from '@/hooks/use-auth';

vi.mock('@/hooks/use-auth', () => ({
  useAuth: vi.fn(),
}));

vi.mock('@/components/ui/dropdown-menu', () => ({
  DropdownMenu: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  DropdownMenuTrigger: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  DropdownMenuContent: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  DropdownMenuItem: ({ children, onSelect, asChild }: { children: ReactNode; onSelect?: () => void; asChild?: boolean }) =>
    asChild ? <div>{children}</div> : <button onClick={onSelect}>{children}</button>,
  DropdownMenuSeparator: () => <hr />,
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, ...props }: ComponentProps<'a'>) => <a {...props}>{children}</a>,
}));

const mockUseAuth = vi.mocked(useAuth);

const renderHeaderUser = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: {
        retry: false,
      },
      queries: {
        retry: false,
      },
    },
  });

  render(
    <QueryClientProvider client={queryClient}>
      <HeaderUser />
    </QueryClientProvider>,
  );
};

afterEach(() => {
  document.body.innerHTML = '';
  vi.clearAllMocks();
});

describe('HeaderUser', () => {
  it('opens the change password dialog from the avatar menu', async () => {
    mockUseAuth.mockReturnValue({
      authStatus: 'authenticated',
      authenticate: vi.fn(),
      isAuthenticated: true,
      logout: vi.fn(),
      user: {
        id: 'user-1',
        email: 'gabriel@example.com',
        name: 'Gabriel',
      },
    });

    renderHeaderUser();

    fireEvent.click(screen.getByText('Change password'));

    expect(await screen.findByText('Update your password without leaving the current page.')).toBeTruthy();
  });
});

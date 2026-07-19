// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ChangePasswordDialog } from './change-password-dialog';
import type { ReactNode } from 'react';
import { changePassword } from '@/features/auth/services/users';
import { handleSuccess } from '@/shared/utils/handle-success';

vi.mock('@/features/auth/services/users', () => ({
  changePassword: vi.fn(),
}));

vi.mock('@/shared/utils/handle-success', () => ({
  handleSuccess: vi.fn(),
}));

const mockChangePassword = vi.mocked(changePassword);
const mockHandleSuccess = vi.mocked(handleSuccess);

const createWrapper = () => {
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

  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

const renderDialog = (onOpenChange = vi.fn()) => {
  render(<ChangePasswordDialog open onOpenChange={onOpenChange} />, {
    wrapper: createWrapper(),
  });

  return { onOpenChange };
};

afterEach(() => {
  document.body.innerHTML = '';
  vi.clearAllMocks();
});

describe('ChangePasswordDialog', () => {
  it('validates required fields, password length, and confirmation match', async () => {
    renderDialog();

    fireEvent.change(screen.getByLabelText('Current password'), { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText('New password'), { target: { value: '123' } });
    fireEvent.change(screen.getByLabelText('Confirm new password'), { target: { value: '456' } });
    fireEvent.click(screen.getByRole('button', { name: 'Update password' }));

    expect(await screen.findByText('password must be at least 6 characters')).toBeTruthy();
    expect(screen.getByText('new password confirmation must match new password')).toBeTruthy();
    expect(mockChangePassword).not.toHaveBeenCalled();
  });

  it('submits the password change and closes the dialog on success', async () => {
    mockChangePassword.mockResolvedValue(undefined);
    const { onOpenChange } = renderDialog();

    fireEvent.change(screen.getByLabelText('Current password'), { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText('New password'), { target: { value: 'newpassword123' } });
    fireEvent.change(screen.getByLabelText('Confirm new password'), { target: { value: 'newpassword123' } });
    fireEvent.click(screen.getByRole('button', { name: 'Update password' }));

    await waitFor(() => {
      expect(mockChangePassword).toHaveBeenCalledWith({
        old_password: 'password123',
        new_password: 'newpassword123',
        new_password_confirmation: 'newpassword123',
      });
    });
    expect(mockHandleSuccess).toHaveBeenCalledWith('Password updated');
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('shows the API error without closing the dialog when the request fails', async () => {
    mockChangePassword.mockRejectedValue(new Error('old password is incorrect'));
    const { onOpenChange } = renderDialog();

    fireEvent.change(screen.getByLabelText('Current password'), { target: { value: 'wrongpassword' } });
    fireEvent.change(screen.getByLabelText('New password'), { target: { value: 'newpassword123' } });
    fireEvent.change(screen.getByLabelText('Confirm new password'), { target: { value: 'newpassword123' } });
    fireEvent.click(screen.getByRole('button', { name: 'Update password' }));

    expect(await screen.findByText('old password is incorrect')).toBeTruthy();
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
    expect(mockHandleSuccess).not.toHaveBeenCalled();
  });
});

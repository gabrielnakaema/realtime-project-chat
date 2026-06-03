// @vitest-environment jsdom

import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { MCPAPIKeyForm } from './mcp-key-form';
import { useMCPAPIAvailableScopes } from '@/hooks/use-mcp-api-available-scopes';

afterEach(() => {
  document.body.innerHTML = '';
});

vi.mock('@/hooks/use-mcp-api-available-scopes', () => ({
  useMCPAPIAvailableScopes: vi.fn(),
}));

const availableScopes = [
  { scope: 'tasks:read', label: 'Read tasks', title: 'Inspect tasks.' },
  { scope: 'tasks:move', label: 'Move tasks', title: 'Move tasks across columns.' },
];

const mockUseMCPAPIAvailableScopes = vi.mocked(useMCPAPIAvailableScopes);

describe('MCPAPIKeyForm', () => {
  beforeEach(() => {
    mockUseMCPAPIAvailableScopes.mockReturnValue({
      availableScopes,
      isLoadingScopes: false,
      isScopeLoadFailed: false,
      refetchScopes: vi.fn(),
    });
  });

  it('preloads name and scopes from the provided key data', () => {
    render(
      <MCPAPIKeyForm
        defaultValues={{ name: 'Editor agent', scopes: ['tasks:move'] }}
        isSubmitting={false}
        submitLabel="Save changes"
        onCancel={() => {}}
        onSubmit={() => {}}
      />,
    );

    const [readTasksCheckbox, moveTasksCheckbox] = screen.getAllByRole<HTMLInputElement>('checkbox');
    const keyNameInput = screen.getByLabelText<HTMLInputElement>('Key name');

    expect(keyNameInput.value).toBe('Editor agent');
    expect(moveTasksCheckbox.checked).toBe(true);
    expect(readTasksCheckbox.checked).toBe(false);
  });

  it('notifies callers when the form changes', () => {
    const onFormChange = vi.fn();

    render(
      <MCPAPIKeyForm
        defaultValues={{ name: 'Reader', scopes: ['tasks:read'] }}
        isSubmitting={false}
        submitLabel="Save changes"
        onCancel={() => {}}
        onFormChange={onFormChange}
        onSubmit={() => {}}
      />,
    );

    const [, moveTasksCheckbox] = screen.getAllByRole<HTMLInputElement>('checkbox');
    const keyNameInput = screen.getByLabelText<HTMLInputElement>('Key name');

    fireEvent.change(keyNameInput, { target: { value: 'Board mover' } });
    fireEvent.click(moveTasksCheckbox);

    expect(onFormChange).toHaveBeenCalledTimes(2);
  });

  it('renders a retry state when scopes fail to load', () => {
    const refetchScopes = vi.fn();
    mockUseMCPAPIAvailableScopes.mockReturnValue({
      availableScopes: [],
      isLoadingScopes: false,
      isScopeLoadFailed: true,
      refetchScopes,
    });

    render(
      <MCPAPIKeyForm
        defaultValues={{ name: '', scopes: [] }}
        isSubmitting={false}
        submitLabel="Save changes"
        onCancel={() => {}}
        onSubmit={() => {}}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Retry' }));

    expect(refetchScopes).toHaveBeenCalledTimes(1);
  });
});

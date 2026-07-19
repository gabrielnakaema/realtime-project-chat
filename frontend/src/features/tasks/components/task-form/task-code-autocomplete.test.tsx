// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { FormProvider, useForm } from 'react-hook-form';
import { TaskCodeAutocomplete } from './task-code-autocomplete';
import type { ReactNode } from 'react';
import type { ITaskForm } from '@/features/tasks/schemas/task.schema';
import { suggestTaskCodes } from '@/features/tasks/services/tasks';

vi.mock('@/features/tasks/services/tasks', () => ({
  suggestTaskCodes: vi.fn(),
}));

const mockSuggestTaskCodes = vi.mocked(suggestTaskCodes);

class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}

globalThis.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver;

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

function TestForm() {
  const form = useForm<ITaskForm>({
    defaultValues: {
      title: '',
      description: '',
      priority: '',
      code: '',
      depends_on_task_ids: [],
    },
  });

  return (
    <FormProvider {...form}>
      <TaskCodeAutocomplete projectId="project-1" />
    </FormProvider>
  );
}

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('TaskCodeAutocomplete', () => {
  it('keeps the field compact and the suggestions open on the initial click', () => {
    render(<TestForm />, { wrapper: createWrapper() });

    const input = screen.getByRole<HTMLInputElement>('combobox', { name: 'Code' });
    const command = input.closest('[data-slot="command"]');
    const popoverAnchor = input.closest('[data-slot="popover-anchor"]');

    expect(command?.className).toContain('h-auto');
    expect(popoverAnchor?.className).toContain('w-full');

    fireEvent.focus(input);
    fireEvent.pointerDown(input);

    expect(screen.getByText('Type at least 2 characters.')).toBeTruthy();
    expect(input.getAttribute('aria-expanded')).toBe('true');
  });

  it('closes with Escape and reopens with ArrowDown', () => {
    render(<TestForm />, { wrapper: createWrapper() });

    const input = screen.getByRole<HTMLInputElement>('combobox', { name: 'Code' });
    fireEvent.focus(input);
    fireEvent.keyDown(input, { key: 'Escape' });

    expect(screen.queryByText('Type at least 2 characters.')).toBeNull();

    fireEvent.keyDown(input, { key: 'ArrowDown' });

    expect(screen.getByText('Type at least 2 characters.')).toBeTruthy();
  });

  it('loads code suggestions and writes the selected code into the form field', async () => {
    mockSuggestTaskCodes.mockResolvedValue([
      { code: 'BACKEND-3', kind: 'next' },
      { code: 'BACKEND-2', kind: 'existing' },
    ]);

    render(<TestForm />, { wrapper: createWrapper() });

    const input = screen.getByLabelText<HTMLInputElement>('Code');
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: 'BACKEND-' } });

    await waitFor(() => {
      expect(mockSuggestTaskCodes).toHaveBeenCalledWith({
        projectId: 'project-1',
        prefix: 'BACKEND-',
        limit: 8,
      });
    });

    fireEvent.click(await screen.findByText('BACKEND-3'));

    expect(input.value).toBe('BACKEND-3');
  });

  it('selects a suggestion with the keyboard', async () => {
    mockSuggestTaskCodes.mockResolvedValue([
      { code: 'BACKEND-3', kind: 'next' },
      { code: 'BACKEND-2', kind: 'existing' },
    ]);

    render(<TestForm />, { wrapper: createWrapper() });

    const input = screen.getByRole<HTMLInputElement>('combobox', { name: 'Code' });
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: 'BACKEND-' } });

    await screen.findByText('BACKEND-3');
    fireEvent.keyDown(input, { key: 'Enter' });

    expect(input.value).toBe('BACKEND-3');
  });

  it('shows a retry action when loading suggestions fails', async () => {
    mockSuggestTaskCodes.mockRejectedValueOnce(new Error('network unavailable')).mockResolvedValueOnce([]);

    render(<TestForm />, { wrapper: createWrapper() });

    const input = screen.getByRole<HTMLInputElement>('combobox', { name: 'Code' });
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: 'BACKEND-' } });

    expect(await screen.findByText('Could not load code suggestions.')).toBeTruthy();
    expect(screen.queryByText('No matching codes found.')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    await waitFor(() => expect(mockSuggestTaskCodes).toHaveBeenCalledTimes(2));
    expect(await screen.findByText('No matching codes found.')).toBeTruthy();
  });

  it('hides stale suggestions while a changed prefix is debouncing', async () => {
    mockSuggestTaskCodes.mockResolvedValue([{ code: 'BACKEND-3', kind: 'next' }]);

    render(<TestForm />, { wrapper: createWrapper() });

    const input = screen.getByRole<HTMLInputElement>('combobox', { name: 'Code' });
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: 'BACKEND-' } });

    await screen.findByText('BACKEND-3');
    fireEvent.change(input, { target: { value: 'FRONTEND-' } });

    expect(screen.queryByText('BACKEND-3')).toBeNull();
  });
});

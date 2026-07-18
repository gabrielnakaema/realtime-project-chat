// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ColumnsProjectSettings } from './columns-project-settings';
import type { Project } from '@/types/project';
import { getProject, updateProject } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';
import { handleError } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';

vi.mock('@/services/projects', () => ({
  getProject: vi.fn(),
  updateProject: vi.fn(),
}));

vi.mock('@/utils/handle-error', () => ({
  handleError: vi.fn(),
}));

vi.mock('@/utils/handle-success', () => ({
  handleSuccess: vi.fn(),
}));

const mockGetProject = vi.mocked(getProject);
const mockUpdateProject = vi.mocked(updateProject);
const mockHandleError = vi.mocked(handleError);
const mockHandleSuccess = vi.mocked(handleSuccess);

const project: Project = {
  id: 'project-1',
  user_id: 'creator-1',
  name: 'Project Chat',
  description: 'Realtime project planning',
  repository_url: 'https://github.com/acme/project-chat',
  repository_owner: 'acme',
  repository_name: 'project-chat',
  default_branch: 'main',
  branch_name_prefix: 'task/',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  members: [],
  columns: [
    {
      id: 'column-pending',
      project_id: 'project-1',
      name: 'Pending',
      description: 'Ready to start',
      color: '#64748B',
      position: 0,
      is_done_column: false,
    },
    {
      id: 'column-doing',
      project_id: 'project-1',
      name: 'Doing',
      description: 'In progress',
      color: '#2563EB',
      position: 1,
      is_done_column: false,
    },
    {
      id: 'column-done',
      project_id: 'project-1',
      name: 'Done',
      description: 'Completed work',
      color: '#059669',
      position: 2,
      is_done_column: true,
    },
  ],
};

const renderSettings = () => {
  mockGetProject.mockResolvedValue(project);
  mockUpdateProject.mockResolvedValue(project);

  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false, staleTime: Infinity },
    },
  });
  queryClient.setQueryData(projectQueryKeys.details(project.id), project);
  const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');

  render(
    <QueryClientProvider client={queryClient}>
      <ColumnsProjectSettings projectId={project.id} />
    </QueryClientProvider>,
  );

  return { invalidateQueries };
};

const openColumn = (name: string) => {
  const item = screen
    .getByText(name, { selector: '[data-slot="accordion-trigger"] span.text-foreground' })
    .closest('[data-slot="accordion-item"]');

  if (!(item instanceof HTMLElement)) throw new Error(`Could not find the ${name} column`);

  const trigger = item.querySelector<HTMLButtonElement>('[data-slot="accordion-trigger"]');
  if (!trigger) throw new Error(`Could not find the ${name} accordion trigger`);

  fireEvent.click(trigger);

  return item;
};

afterEach(() => {
  document.body.innerHTML = '';
  vi.resetAllMocks();
});

describe('ColumnsProjectSettings', () => {
  it('edits column fields, changes the done column, and preserves the rest of the project payload', async () => {
    const { invalidateQueries } = renderSettings();

    fireEvent.change(screen.getByLabelText('COLUMN NAME'), { target: { value: 'Backlog' } });

    const doingColumn = openColumn('Doing');
    fireEvent.click(within(doingColumn).getByRole('button', { name: /Done column/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockUpdateProject).toHaveBeenCalledWith({
        id: project.id,
        name: project.name,
        description: project.description,
        repository_url: project.repository_url,
        repository_owner: project.repository_owner,
        repository_name: project.repository_name,
        default_branch: project.default_branch,
        branch_name_prefix: project.branch_name_prefix,
        columns: [
          {
            id: 'column-pending',
            name: 'Backlog',
            description: 'Ready to start',
            color: '#64748B',
            is_done_column: false,
          },
          {
            id: 'column-doing',
            name: 'Doing',
            description: 'In progress',
            color: '#2563EB',
            is_done_column: true,
          },
          {
            id: 'column-done',
            name: 'Done',
            description: 'Completed work',
            color: '#059669',
            is_done_column: false,
          },
        ],
        deleted_columns: [],
      });
    });
    expect(mockHandleSuccess).toHaveBeenCalledWith('Project columns saved successfully');
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['projects'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['tasks'] });
  });

  it('reorders from the accordion header without collapsing the expanded column', async () => {
    renderSettings();

    const pendingItem = screen
      .getByText('Pending', { selector: '[data-slot="accordion-trigger"] span.text-foreground' })
      .closest('[data-slot="accordion-item"]');
    if (!(pendingItem instanceof HTMLElement)) throw new Error('Could not find the Pending column');

    const pendingTrigger = pendingItem.querySelector('[data-slot="accordion-trigger"]');
    expect(pendingTrigger?.getAttribute('aria-expanded')).toBe('true');

    fireEvent.click(within(pendingItem).getByRole('button', { name: 'Move Pending down' }));

    expect(
      screen
        .getAllByText(/Pending|Doing|Done/, { selector: '[data-slot="accordion-trigger"] span.text-foreground' })
        .map((item) => item.textContent),
    ).toEqual(['Doing', 'Pending', 'Done']);

    const movedPendingItem = screen
      .getByText('Pending', { selector: '[data-slot="accordion-trigger"] span.text-foreground' })
      .closest('[data-slot="accordion-item"]');
    expect(movedPendingItem?.querySelector('[data-slot="accordion-trigger"]')?.getAttribute('aria-expanded')).toBe(
      'true',
    );

    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockUpdateProject).toHaveBeenCalledWith(
        expect.objectContaining({
          columns: [
            expect.objectContaining({ id: 'column-doing' }),
            expect.objectContaining({ id: 'column-pending' }),
            expect.objectContaining({ id: 'column-done' }),
          ],
        }),
      );
    });
  });

  it('adds a column and validates its existing schema fields before saving', async () => {
    renderSettings();

    fireEvent.click(screen.getByRole('button', { name: 'Add column' }));

    expect(screen.getByText('COLUMNS • 4')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    expect(await screen.findByText('Column name is required')).toBeTruthy();
    expect(mockUpdateProject).not.toHaveBeenCalled();

    const newColumnName = document.querySelector<HTMLInputElement>('#column-name-3');
    if (!newColumnName) throw new Error('New column name input was not rendered');

    fireEvent.change(newColumnName, { target: { value: 'Review' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockUpdateProject).toHaveBeenCalledWith(
        expect.objectContaining({
          columns: expect.arrayContaining([
            {
              name: 'Review',
              description: '',
              color: '#D97706',
              is_done_column: false,
            },
          ]),
          deleted_columns: [],
        }),
      );
    });
  });

  it('removes a saved column with a valid task reassignment before saving', async () => {
    renderSettings();

    const doingColumn = openColumn('Doing');
    fireEvent.click(within(doingColumn).getByRole('button', { name: 'Delete Doing' }));

    expect(screen.getByText('COLUMNS • 2')).toBeTruthy();
    expect(screen.getByText('PENDING REMOVALS')).toBeTruthy();
    expect(screen.getByText('Doing', { selector: 'strong' })).toBeTruthy();
    expect(screen.getByText(/column will be deleted after tasks are moved\./)).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockUpdateProject).toHaveBeenCalledWith(
        expect.objectContaining({
          columns: [expect.objectContaining({ id: 'column-pending' }), expect.objectContaining({ id: 'column-done' })],
          deleted_columns: [
            {
              id: 'column-doing',
              move_tasks_to_column_id: 'column-pending',
            },
          ],
        }),
      );
    });
  });

  it('restores a pending column removal without submitting a deletion', async () => {
    renderSettings();

    const doingColumn = openColumn('Doing');
    fireEvent.click(within(doingColumn).getByRole('button', { name: 'Delete Doing' }));
    fireEvent.click(screen.getByRole('button', { name: 'Restore' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockUpdateProject).toHaveBeenCalledWith(
        expect.objectContaining({
          columns: expect.arrayContaining([expect.objectContaining({ id: 'column-doing', name: 'Doing' })]),
          deleted_columns: [],
        }),
      );
    });
  });

  it('keeps unsaved changes and reports the API error when saving fails', async () => {
    const error = new Error('Unable to update columns');
    renderSettings();
    mockUpdateProject.mockRejectedValue(error);

    const nameInput = screen.getByLabelText<HTMLInputElement>('COLUMN NAME');
    fireEvent.change(nameInput, { target: { value: 'Backlog' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockHandleError).toHaveBeenCalledWith(error);
    });
    expect(nameInput.value).toBe('Backlog');
    expect(screen.getByText('You have unsaved changes.')).toBeTruthy();
    expect(mockHandleSuccess).not.toHaveBeenCalled();
  });
});

// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { GeneralProjectSettingsForm } from './general-project-settings-form';
import type { Project } from '@/features/projects/types/project';
import { deleteProject, getProject, updateProject } from '@/features/projects/services/projects';
import { projectQueryKeys } from '@/shared/services/query-keys';
import { handleError } from '@/shared/utils/handle-error';
import { handleSuccess } from '@/shared/utils/handle-success';

const mockNavigate = vi.fn();

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('@/shared/components/text-editor', () => ({
  TextEditor: ({
    error,
    id,
    initialValue,
    label,
    onChange,
  }: {
    error?: string;
    id?: string;
    initialValue: string;
    label?: string;
    onChange: (value: string) => void;
  }) => {
    const [value, setValue] = useState(initialValue);

    return (
      <div>
        <label htmlFor={id}>{label}</label>
        <textarea
          id={id}
          value={value}
          onChange={(event) => {
            setValue(event.target.value);
            onChange(event.target.value);
          }}
        />
        {error && <p>{error}</p>}
      </div>
    );
  },
}));

vi.mock('@/features/projects/services/projects', () => ({
  getProject: vi.fn(),
  updateProject: vi.fn(),
  deleteProject: vi.fn(),
}));

vi.mock('@/shared/utils/handle-error', () => ({
  handleError: vi.fn(),
}));

vi.mock('@/shared/utils/handle-success', () => ({
  handleSuccess: vi.fn(),
}));

const mockGetProject = vi.mocked(getProject);
const mockUpdateProject = vi.mocked(updateProject);
const mockDeleteProject = vi.mocked(deleteProject);
const mockHandleError = vi.mocked(handleError);
const mockHandleSuccess = vi.mocked(handleSuccess);

const project: Project = {
  id: 'project-1',
  user_id: 'creator-1',
  name: 'Project Chat',
  description: '<p>Realtime project planning</p>',
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
      id: 'column-done',
      project_id: 'project-1',
      name: 'Done',
      description: 'Completed work',
      color: '#059669',
      position: 1,
      is_done_column: true,
    },
  ],
};

const renderSettings = (updatedProject: Project = project, preloadProject = true) => {
  mockGetProject.mockResolvedValue(updatedProject);
  mockUpdateProject.mockResolvedValue(updatedProject);

  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false, staleTime: Infinity },
    },
  });
  if (preloadProject) {
    queryClient.setQueryData(projectQueryKeys.details(project.id), project);
  }
  const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');

  render(
    <QueryClientProvider client={queryClient}>
      <GeneralProjectSettingsForm projectId={project.id} />
    </QueryClientProvider>,
  );

  return { invalidateQueries };
};

afterEach(() => {
  document.body.innerHTML = '';
  vi.resetAllMocks();
});

describe('GeneralProjectSettingsForm', () => {
  it('fills the project description after the initial project request resolves', async () => {
    renderSettings(project, false);

    expect(await screen.findByLabelText<HTMLTextAreaElement>('PROJECT DESCRIPTION')).toHaveProperty(
      'value',
      project.description,
    );
  });

  it('saves general settings while preserving the existing project columns', async () => {
    const submittedDescription = '<p>Updated description draft</p>';
    const updatedProject: Project = {
      ...project,
      name: 'Updated Project Chat',
      description: '<p>Updated description from server</p>',
      repository_url: 'https://github.com/example/updated-project',
      repository_owner: 'example',
      repository_name: 'updated-project',
      default_branch: 'develop',
      branch_name_prefix: 'feature/',
    };
    const { invalidateQueries } = renderSettings(updatedProject);

    fireEvent.change(screen.getByLabelText('PROJECT NAME'), { target: { value: updatedProject.name } });
    fireEvent.change(screen.getByLabelText('PROJECT DESCRIPTION'), {
      target: { value: submittedDescription },
    });
    fireEvent.change(screen.getByLabelText('URL'), { target: { value: updatedProject.repository_url } });
    fireEvent.change(screen.getByLabelText('OWNER'), { target: { value: updatedProject.repository_owner } });
    fireEvent.change(screen.getByLabelText('NAME'), { target: { value: updatedProject.repository_name } });
    fireEvent.change(screen.getByLabelText('DEFAULT BRANCH'), { target: { value: updatedProject.default_branch } });
    fireEvent.change(screen.getByLabelText('BRANCH NAME PREFIX'), {
      target: { value: updatedProject.branch_name_prefix },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockUpdateProject).toHaveBeenCalledWith({
        id: project.id,
        name: updatedProject.name,
        description: submittedDescription,
        repository_url: updatedProject.repository_url,
        repository_owner: updatedProject.repository_owner,
        repository_name: updatedProject.repository_name,
        default_branch: updatedProject.default_branch,
        branch_name_prefix: updatedProject.branch_name_prefix,
        columns: [
          {
            id: 'column-pending',
            name: 'Pending',
            description: 'Ready to start',
            color: '#64748B',
            is_done_column: false,
          },
          {
            id: 'column-done',
            name: 'Done',
            description: 'Completed work',
            color: '#059669',
            is_done_column: true,
          },
        ],
        deleted_columns: [],
      });
    });
    expect(mockHandleSuccess).toHaveBeenCalledWith('Project saved successfully');
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['projects'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['tasks'] });
    await waitFor(() => {
      expect(screen.queryByText('You have unsaved changes.')).toBeNull();
      expect(screen.getByLabelText<HTMLTextAreaElement>('PROJECT DESCRIPTION').value).toBe(updatedProject.description);
    });
  });

  it('keeps unsaved general settings and reports the API error when saving fails', async () => {
    const error = new Error('Unable to update project');
    renderSettings();
    mockUpdateProject.mockRejectedValue(error);
    const nameInput = screen.getByLabelText<HTMLInputElement>('PROJECT NAME');

    fireEvent.change(nameInput, { target: { value: 'Unsaved project name' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    await waitFor(() => {
      expect(mockHandleError).toHaveBeenCalledWith(error);
    });
    expect(nameInput.value).toBe('Unsaved project name');
    expect(screen.getByText('You have unsaved changes.')).toBeTruthy();
    expect(mockHandleSuccess).not.toHaveBeenCalled();
  });

  it('deletes the project and navigates away after confirming', async () => {
    renderSettings();
    mockDeleteProject.mockResolvedValue(true);

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Delete project' }));

    await waitFor(() => {
      expect(mockDeleteProject).toHaveBeenCalledWith(project.id);
    });
    expect(mockHandleSuccess).toHaveBeenCalledWith('Project deleted successfully');
    expect(mockNavigate).toHaveBeenCalledWith({ to: '/projects' });
  });

  it('shows a pending state on the confirm button while deleting', async () => {
    renderSettings();
    let resolveDelete: (value: boolean) => void = () => {};
    mockDeleteProject.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveDelete = resolve;
        }),
    );

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Delete project' }));

    expect(await screen.findByRole('button', { name: 'Deleting...' })).toBeTruthy();

    resolveDelete(true);
    await waitFor(() => {
      expect(mockHandleSuccess).toHaveBeenCalledWith('Project deleted successfully');
    });
  });

  it('shows an error toast and keeps the dialog usable when deletion fails', async () => {
    renderSettings();
    const error = new Error('Unable to delete project');
    mockDeleteProject.mockRejectedValue(error);

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Delete project' }));

    await waitFor(() => {
      expect(mockHandleError).toHaveBeenCalledWith(error);
    });
    expect(mockNavigate).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Delete project' })).toBeTruthy();
  });
});

// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { MembersProjectSettings } from './members-project-settings';
import type { Project } from '@/types/project';
import { ProjectMemberRole } from '@/types/project';
import { useAuth } from '@/hooks/use-auth';
import { createProjectMember, getProject, removeProjectMember } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';
import { handleError } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';

vi.mock('@/hooks/use-auth', () => ({
  useAuth: vi.fn(),
}));

vi.mock('@/services/projects', () => ({
  createProjectMember: vi.fn(),
  getProject: vi.fn(),
  removeProjectMember: vi.fn(),
}));

vi.mock('@/utils/handle-error', () => ({
  handleError: vi.fn(),
}));

vi.mock('@/utils/handle-success', () => ({
  handleSuccess: vi.fn(),
}));

const mockUseAuth = vi.mocked(useAuth);
const mockCreateProjectMember = vi.mocked(createProjectMember);
const mockGetProject = vi.mocked(getProject);
const mockRemoveProjectMember = vi.mocked(removeProjectMember);
const mockHandleError = vi.mocked(handleError);
const mockHandleSuccess = vi.mocked(handleSuccess);

const project: Project = {
  id: 'project-1',
  user_id: 'creator-1',
  name: 'Project Chat',
  description: '',
  repository_url: '',
  repository_owner: '',
  repository_name: '',
  default_branch: 'main',
  branch_name_prefix: 'task/',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  columns: [],
  members: [
    {
      id: 'membership-1',
      project_id: 'project-1',
      user_id: 'creator-1',
      role: ProjectMemberRole.Creator,
      user: {
        id: 'creator-1',
        name: 'Project Creator',
        email: 'creator@example.com',
      },
    },
    {
      id: 'membership-2',
      project_id: 'project-1',
      user_id: 'member-1',
      role: ProjectMemberRole.Member,
      user: {
        id: 'member-1',
        name: 'Project Member',
        email: 'member@example.com',
      },
    },
  ],
};

const setAuthenticatedUser = (userId: string) => {
  mockUseAuth.mockReturnValue({
    authStatus: 'authenticated',
    authenticate: vi.fn(),
    isAuthenticated: true,
    logout: vi.fn(),
    user: {
      id: userId,
      name: 'Current User',
      email: 'current@example.com',
    },
  });
};

const renderSettings = () => {
  mockGetProject.mockResolvedValue(project);
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
      <MembersProjectSettings projectId="project-1" />
    </QueryClientProvider>,
  );

  return { invalidateQueries };
};

afterEach(() => {
  document.body.innerHTML = '';
  vi.resetAllMocks();
});

describe('MembersProjectSettings', () => {
  it('lets the project creator choose a non-creator for removal', () => {
    setAuthenticatedUser('creator-1');

    renderSettings();

    expect(screen.queryByLabelText('Remove Project Creator from project')).toBeNull();
    fireEvent.click(screen.getByLabelText('Remove Project Member from project'));

    expect(screen.getByRole('heading', { name: 'Remove member from project?' })).toBeTruthy();
    expect(screen.getByText('Are you sure you want to remove Project Member from the project?')).toBeTruthy();
  });

  it('does not offer member removal to a non-creator', () => {
    setAuthenticatedUser('member-1');

    renderSettings();

    expect(screen.queryByRole('button', { name: /remove .* from project/i })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Add member' })).toBeNull();
  });

  it('validates the invite email before submitting', async () => {
    setAuthenticatedUser('creator-1');
    renderSettings();

    fireEvent.click(screen.getByRole('button', { name: 'Add member' }));

    expect(await screen.findByText('Email is required.')).toBeTruthy();
    expect(mockCreateProjectMember).not.toHaveBeenCalled();

    fireEvent.change(screen.getByLabelText('Email address'), { target: { value: 'not-an-email' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add member' }));

    expect(await screen.findByText('Enter a valid email address.')).toBeTruthy();
    expect(mockCreateProjectMember).not.toHaveBeenCalled();
  });

  it('adds a member, resets the form, and refreshes project data', async () => {
    setAuthenticatedUser('creator-1');
    mockCreateProjectMember.mockResolvedValue({});
    const { invalidateQueries } = renderSettings();
    const emailInput = screen.getByLabelText<HTMLInputElement>('Email address');

    fireEvent.change(emailInput, { target: { value: 'new-member@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add member' }));

    await waitFor(() => {
      expect(mockCreateProjectMember).toHaveBeenCalledWith({
        projectId: 'project-1',
        email: 'new-member@example.com',
      });
    });
    expect(mockHandleSuccess).toHaveBeenCalledWith('Member added successfully');
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['projects'] });
    await waitFor(() => {
      expect(emailInput.value).toBe('');
    });
  });

  it('keeps the invite email and reports an API error when adding a member fails', async () => {
    const error = new Error('User not found');
    setAuthenticatedUser('creator-1');
    mockCreateProjectMember.mockRejectedValue(error);
    renderSettings();
    const emailInput = screen.getByLabelText<HTMLInputElement>('Email address');

    fireEvent.change(emailInput, { target: { value: 'missing@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add member' }));

    await waitFor(() => {
      expect(mockHandleError).toHaveBeenCalledWith(error);
    });
    expect(emailInput.value).toBe('missing@example.com');
    expect(mockHandleSuccess).not.toHaveBeenCalled();
  });

  it('removes the selected member, closes the dialog, and invalidates project board data', async () => {
    setAuthenticatedUser('creator-1');
    mockRemoveProjectMember.mockResolvedValue(true);
    const { invalidateQueries } = renderSettings();

    fireEvent.click(screen.getByLabelText('Remove Project Member from project'));
    fireEvent.click(screen.getByRole('button', { name: 'Remove member' }));

    await waitFor(() => {
      expect(mockRemoveProjectMember).toHaveBeenCalledWith({
        projectId: 'project-1',
        memberUserId: 'member-1',
      });
    });
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Remove member from project?' })).toBeNull();
    });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['projects'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['tasks'] });
  });

  it('keeps the confirmation open and reports an API error when removal fails', async () => {
    const error = new Error('Unable to remove member');
    setAuthenticatedUser('creator-1');
    mockRemoveProjectMember.mockRejectedValue(error);
    renderSettings();

    fireEvent.click(screen.getByLabelText('Remove Project Member from project'));
    fireEvent.click(screen.getByRole('button', { name: 'Remove member' }));

    await waitFor(() => {
      expect(mockHandleError).toHaveBeenCalledWith(error);
    });
    expect(screen.getByRole('heading', { name: 'Remove member from project?' })).toBeTruthy();
  });
});

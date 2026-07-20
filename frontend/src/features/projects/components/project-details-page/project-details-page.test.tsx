// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Project } from '@/features/projects/types/project';
import { ProjectDetailsPage } from '@/routes/_protected/projects/$projectId/index';
import { getProject } from '@/features/projects/services/projects';

vi.mock('@/features/projects/services/projects', () => ({
  getProject: vi.fn(),
}));

vi.mock('@/features/tasks/hooks/use-task-details-routing', () => ({
  useTaskDetailsRouting: () => ({
    selectedTaskId: undefined,
    selectedCommentId: undefined,
    selectedCommentCreatedAt: undefined,
    isEditingTask: false,
    closeTask: vi.fn(),
    startEditingTask: vi.fn(),
    stopEditingTask: vi.fn(),
  }),
}));

vi.mock('@/features/projects/components/project-details-page/project-details-header', () => ({
  ProjectDetailsHeader: ({ project }: { project: Project }) => <div>Header: {project.name}</div>,
}));

vi.mock('@/features/tasks/components/kanban-board', () => ({
  KanbanBoard: ({ project }: { project: Project }) => <div>Board: {project.name}</div>,
}));

vi.mock('@/features/tasks/components/task-details', () => ({
  TaskDetails: () => null,
}));

vi.mock('@/features/tasks/components/task-form/edit-task', () => ({
  EditTask: () => null,
}));

const mockGetProject = vi.mocked(getProject);

const project: Project = {
  id: 'project-1',
  user_id: 'user-1',
  name: 'Project Chat',
  description: 'Realtime project planning',
  repository_url: '',
  repository_owner: '',
  repository_name: '',
  default_branch: 'main',
  branch_name_prefix: 'task/',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  members: [],
  columns: [],
};

const renderProjectDetails = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <ProjectDetailsPage projectId={project.id} />
    </QueryClientProvider>,
  );
};

describe('ProjectDetailsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('renders a project-shaped skeleton while project details load', () => {
    mockGetProject.mockImplementation(() => new Promise<Project>(() => undefined));

    renderProjectDetails();

    expect(screen.getByLabelText('Loading project')).toBeTruthy();
    expect(screen.queryByText(/Header:/)).toBeNull();
  });

  it('renders the project header and board after loading', async () => {
    mockGetProject.mockResolvedValue(project);

    renderProjectDetails();

    expect(await screen.findByText('Header: Project Chat')).toBeTruthy();
    expect(screen.getByText('Board: Project Chat')).toBeTruthy();
    expect(screen.queryByLabelText('Loading project')).toBeNull();
  });

  it('shows an error state and retries the project request', async () => {
    mockGetProject.mockRejectedValueOnce(new Error('Network error')).mockResolvedValueOnce(project);

    renderProjectDetails();

    expect(await screen.findByText('Project could not be loaded.')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    await waitFor(() => expect(mockGetProject).toHaveBeenCalledTimes(2));
    expect(await screen.findByText('Header: Project Chat')).toBeTruthy();
  });
});

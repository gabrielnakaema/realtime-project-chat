// @vitest-environment jsdom

import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useTaskComments } from './use-task-comments';
import { TaskComments } from '.';
import { useAuth } from '@/hooks/use-auth';

vi.mock('@/hooks/use-auth', () => ({
  useAuth: vi.fn(),
}));

vi.mock('./use-task-comments', () => ({
  useTaskComments: vi.fn(),
}));

vi.mock('../text-editor', () => ({
  TextEditor: ({ onChange, placeholder }: { onChange: (value: string) => void; placeholder?: string }) => (
    <textarea
      aria-label={placeholder ?? 'Comment editor'}
      placeholder={placeholder}
      onChange={(event) => onChange(event.currentTarget.value)}
    />
  ),
}));

vi.mock('../ui/scroll-area', () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

const mockUseAuth = vi.mocked(useAuth);
const mockUseTaskComments = vi.mocked(useTaskComments);
const submitComment = vi.fn();

afterEach(() => {
  document.body.innerHTML = '';
});

beforeEach(() => {
  submitComment.mockReset();
  submitComment.mockResolvedValue(undefined);
});

const baseHookValue = {
  comments: [],
  isLoading: false,
  isSubmitting: false,
  hasPreviousPage: false,
  isFetchingNextPage: false,
  isFetchingPreviousPage: false,
  fetchPreviousPage: vi.fn(),
  sentinelRef: vi.fn(),
  submitComment,
};

describe('TaskComments', () => {
  it('renders AI Agent and You badges for current-user MCP comments', () => {
    mockUseAuth.mockReturnValue({
      user: { id: 'user-1', name: 'Alex', email: 'alex@example.com' },
      isAuthenticated: true,
      logout: vi.fn(),
      authenticate: vi.fn(),
      authStatus: 'authenticated',
    });

    mockUseTaskComments.mockReturnValue({
      ...baseHookValue,
      comments: [
        {
          id: 'comment-1',
          task: null,
          user: { id: 'user-1', name: 'Alex', email: 'alex@example.com' },
          content: '<p>Created via MCP</p>',
          parent_comment_id: null,
          action_origin: 'mcp_agent',
          created_at: '2026-06-03T10:00:00.000Z',
          updated_at: '2026-06-03T10:00:00.000Z',
          replies: [],
        },
      ],
    });

    render(<TaskComments taskId="task-1" projectId="project-1" open />);

    expect(screen.getByText('AI Agent')).not.toBeNull();
    expect(screen.getByText('You')).not.toBeNull();
  });

  it('skips the AI badge when comment origin is missing', () => {
    mockUseAuth.mockReturnValue({
      user: { id: 'user-2', name: 'Jamie', email: 'jamie@example.com' },
      isAuthenticated: true,
      logout: vi.fn(),
      authenticate: vi.fn(),
      authStatus: 'authenticated',
    });

    mockUseTaskComments.mockReturnValue({
      ...baseHookValue,
      comments: [
        {
          id: 'comment-2',
          task: null,
          user: { id: 'user-3', name: 'Taylor', email: 'taylor@example.com' },
          content: '<p>Hello</p>',
          parent_comment_id: null,
          created_at: '2026-06-03T11:00:00.000Z',
          updated_at: '2026-06-03T11:00:00.000Z',
          replies: [],
        },
      ],
    });

    render(<TaskComments taskId="task-1" projectId="project-1" open />);

    expect(screen.queryByText('AI Agent')).toBeNull();
    expect(screen.getByText('Taylor')).not.toBeNull();
  });

  it('submits a reply from the selected comment using local reply state', async () => {
    mockUseAuth.mockReturnValue({
      user: { id: 'user-2', name: 'Jamie', email: 'jamie@example.com' },
      isAuthenticated: true,
      logout: vi.fn(),
      authenticate: vi.fn(),
      authStatus: 'authenticated',
    });

    mockUseTaskComments.mockReturnValue({
      ...baseHookValue,
      comments: [
        {
          id: 'comment-3',
          task: null,
          user: { id: 'user-3', name: 'Taylor', email: 'taylor@example.com' },
          content: '<p>Hello</p>',
          parent_comment_id: null,
          created_at: '2026-06-03T11:00:00.000Z',
          updated_at: '2026-06-03T11:00:00.000Z',
          replies: [],
        },
      ],
    });

    render(<TaskComments taskId="task-1" projectId="project-1" open />);

    fireEvent.click(screen.getByRole('button', { name: 'Reply' }));
    fireEvent.change(screen.getByPlaceholderText('Reply to Taylor...'), {
      target: { value: 'Following up' },
    });
    fireEvent.click(screen.getAllByRole('button', { name: 'Reply' })[1]);

    await waitFor(() => {
      expect(submitComment).toHaveBeenCalledWith({
        content: 'Following up',
        parentCommentId: 'comment-3',
      });
    });

    await waitFor(() => {
      expect(screen.queryByPlaceholderText('Reply to Taylor...')).toBeNull();
    });
  });
});

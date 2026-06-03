// @vitest-environment jsdom

import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { TaskComments } from './task-comments';
import { useAuth } from '@/hooks/use-auth';
import { useTaskComments } from '@/hooks/use-task-comments';

vi.mock('@/hooks/use-auth', () => ({
  useAuth: vi.fn(),
}));

vi.mock('@/hooks/use-task-comments', () => ({
  useTaskComments: vi.fn(),
}));

vi.mock('../text-editor', () => ({
  TextEditor: ({ onChange }: { onChange: (value: string) => void }) => (
    <textarea aria-label="Comment editor" onChange={(event) => onChange(event.currentTarget.value)} />
  ),
}));

vi.mock('../ui/scroll-area', () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

const mockUseAuth = vi.mocked(useAuth);
const mockUseTaskComments = vi.mocked(useTaskComments);

afterEach(() => {
  document.body.innerHTML = '';
});

const baseHookValue = {
  comments: [],
  isLoading: false,
  isSubmitting: false,
  hasPreviousPage: false,
  hasNextPage: false,
  isFetchingNextPage: false,
  isFetchingPreviousPage: false,
  fetchPreviousPage: vi.fn(),
  sentinelRef: vi.fn(),
  commentDraft: '',
  setCommentDraft: vi.fn(),
  composerKey: 0,
  replyDraft: '',
  setReplyDraft: vi.fn(),
  replyEditorKey: 0,
  replyingToId: null,
  submitComment: vi.fn(),
  startReply: vi.fn(),
  cancelReply: vi.fn(),
  submitReply: vi.fn(),
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
});

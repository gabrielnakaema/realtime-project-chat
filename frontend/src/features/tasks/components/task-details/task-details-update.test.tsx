// @vitest-environment jsdom

import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { TaskDetailsUpdate } from './task-details-update';
import type { TaskUpdate } from '@/features/tasks/types/task';

vi.mock('@/shared/components/avatar', () => ({
  Avatar: ({ name }: { name: string }) => <div>{name}</div>,
}));

afterEach(() => {
  document.body.innerHTML = '';
});

const createUpdate = (actionOrigin?: TaskUpdate['action_origin']): TaskUpdate => ({
  id: 'update-1',
  task_id: 'task-1',
  user_id: 'user-1',
  update_type: 'created',
  action_origin: actionOrigin,
  created_at: '2026-06-03T10:00:00.000Z',
  user: {
    id: 'user-1',
    name: 'Alex',
    email: 'alex@example.com',
  },
  changes: [],
});

describe('TaskDetailsUpdate', () => {
  it('renders both AI Agent and You badges for current-user MCP updates', () => {
    render(<TaskDetailsUpdate update={createUpdate('mcp_agent')} isLast currentUserId="user-1" />);

    expect(screen.getByText('AI Agent')).not.toBeNull();
    expect(screen.getByText('You')).not.toBeNull();
  });

  it('does not render the AI agent badge for unknown origins', () => {
    render(<TaskDetailsUpdate update={createUpdate(undefined)} isLast currentUserId="user-1" />);

    expect(screen.queryByText('AI Agent')).toBeNull();
    expect(screen.getByText('You')).not.toBeNull();
  });
});

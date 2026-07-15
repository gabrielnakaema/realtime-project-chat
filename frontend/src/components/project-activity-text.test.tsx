// @vitest-environment jsdom

import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { ProjectActivityText } from './project-activity-text';
import type { TaskProjectActivity } from '@/types/project-activity';

afterEach(() => {
  document.body.innerHTML = '';
});

const createTaskActivity = (actionOrigin?: TaskProjectActivity['action_origin']): TaskProjectActivity => ({
  id: 'activity-1',
  project: {
    id: 'project-1',
    user_id: 'user-1',
    name: 'Project Atlas',
    description: 'desc',
    repository_url: '',
    repository_owner: '',
    repository_name: '',
    default_branch: '',
    branch_name_prefix: '',
    created_at: '2026-06-03T10:00:00.000Z',
    updated_at: '2026-06-03T10:00:00.000Z',
    members: [],
    columns: [],
  },
  actor: {
    id: 'user-1',
    name: 'Alex',
    email: 'alex@example.com',
  },
  activity_type: 'task.created',
  action_origin: actionOrigin,
  activity_data: {},
  entity_id: 'task-1',
  created_at: '2026-06-03T10:00:00.000Z',
  updated_at: '2026-06-03T10:00:00.000Z',
  entity_type: 'task',
  task: {
    id: 'task-1',
    project_id: 'project-1',
    title: 'Ship MCP badge',
    description: 'desc',
    code: 'MCP-1',
    status: 'doing',
    project_column_id: 'column-1',
    project_column: null,
    created_at: '2026-06-03T10:00:00.000Z',
    updated_at: '2026-06-03T10:00:00.000Z',
    priority: 'medium',
    order: 'a0',
    version: 0,
    responsible_id: null,
    due_date: null,
    done_at: null,
    archived_at: null,
    tags: [],
    author_id: 'user-1',
    author: {
      id: 'user-1',
      name: 'Alex',
      email: 'alex@example.com',
    },
    responsible: null,
    updates: [],
    project: null,
  },
});

describe('ProjectActivityText', () => {
  it('renders the AI agent badge for MCP-origin project activity', () => {
    render(<ProjectActivityText activity={createTaskActivity('mcp_agent')} />);

    expect(screen.getByText('AI Agent')).not.toBeNull();
  });

  it('does not render the AI agent badge for user-origin activity', () => {
    render(<ProjectActivityText activity={createTaskActivity('user')} />);

    expect(screen.queryByText('AI Agent')).toBeNull();
  });
});

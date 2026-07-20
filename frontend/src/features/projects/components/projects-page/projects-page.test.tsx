// @vitest-environment jsdom

import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectSummaryCard } from './projects-grid';
import { getDueStatus } from './upcoming-deadlines';
import type { ComponentProps, ReactNode } from 'react';
import type { Project } from '@/features/projects/types/project';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, params, search: _search, to: _to, ...props }: MockLinkProps) => (
    <a href={`/projects/${params.projectId}`} {...props}>
      {children}
    </a>
  ),
}));

interface MockLinkProps extends Omit<ComponentProps<'a'>, 'href'> {
  children: ReactNode;
  params: { projectId: string };
  search?: unknown;
  to: string;
}

const project: Project = {
  id: 'project-1',
  user_id: 'user-1',
  name: 'Atlas',
  description: '<p>Core commerce platform.</p><script>unsafe()</script>',
  repository_url: '',
  repository_owner: '',
  repository_name: '',
  default_branch: 'main',
  branch_name_prefix: 'task/',
  created_at: '2026-07-17T12:00:00Z',
  updated_at: '2026-07-18T12:00:00Z',
  columns: [
    {
      id: 'column-1',
      name: 'Todo',
      description: '',
      color: '#000000',
      position: 0,
      is_done_column: false,
    },
  ],
  members: [
    {
      id: 'membership-1',
      project_id: 'project-1',
      user_id: 'user-1',
      role: 'creator',
      user: { id: 'user-1', email: 'maya@example.com', name: 'Maya Patel' },
    },
    {
      id: 'membership-2',
      project_id: 'project-1',
      user_id: 'user-2',
      role: 'member',
      user: { id: 'user-2', email: 'dev@example.com', name: 'Dev Osei' },
    },
  ],
};

afterEach(() => {
  document.body.innerHTML = '';
});

describe('ProjectSummaryCard', () => {
  it('renders the template card using only project response fields', () => {
    const { container } = render(<ProjectSummaryCard project={project} />);

    expect(screen.getByRole('heading', { name: 'Atlas' })).toBeTruthy();
    expect(screen.getByText('Core commerce platform.')).toBeTruthy();
    expect(screen.getByText('2')).toBeTruthy();
    expect(screen.getByText('members')).toBeTruthy();
    expect(screen.getByText('1')).toBeTruthy();
    expect(screen.getByText('columns')).toBeTruthy();
    expect(screen.getByLabelText('2 project members')).toBeTruthy();
    expect(container.querySelector('script')).toBeNull();
    expect(screen.queryByText('open')).toBeNull();
    expect(screen.queryByText('claimed')).toBeNull();
    expect(screen.queryByText('active')).toBeNull();
  });
});

describe('getDueStatus', () => {
  const now = new Date('2026-07-18T12:00:00Z');

  it('formats overdue, near, and later deadlines like the supplied design', () => {
    expect(getDueStatus('2026-07-17T12:00:00Z', now)).toEqual({ label: '1d overdue', tone: 'overdue' });
    expect(getDueStatus('2026-07-19T12:00:00Z', now)).toEqual({ label: 'Due tomorrow', tone: 'soon' });
    expect(getDueStatus('2026-07-24T12:00:00Z', now)).toEqual({ label: 'Due in 6d', tone: 'default' });
  });
});

import { describe, expect, it } from 'vitest';
import { formatTaskDependencyLabel, getTaskFormValues } from './task-form-utils';
import type { Task } from '@/features/tasks/types/task';

const createTask = (overrides: Partial<Task> = {}): Task => ({
  id: 'task-1',
  project_id: 'project-1',
  title: 'Implement dependencies',
  description: '<p>Details</p>',
  code: 'FRONTEND-3',
  status: 'pending',
  project_column_id: 'column-1',
  project_column: null,
  created_at: '2026-06-08T00:00:00Z',
  updated_at: '2026-06-08T00:00:00Z',
  priority: 'medium',
  order: '001',
  version: 0,
  responsible_id: null,
  due_date: '2026-06-15T12:00:00Z',
  done_at: null,
  archived_at: null,
  tags: ['frontend', 'forms'],
  depends_on_task_ids: ['task-2', 'task-3'],
  author_id: 'user-1',
  author: {
    id: 'user-1',
    name: 'Alex',
    email: 'alex@example.com',
  },
  responsible: null,
  updates: [],
  project: null,
  ...overrides,
});

describe('task form utils', () => {
  it('formats dependency labels with and without codes', () => {
    expect(formatTaskDependencyLabel({ title: 'Write tests', code: 'FRONTEND-2' })).toBe('FRONTEND-2 — Write tests');
    expect(formatTaskDependencyLabel({ title: 'Write tests', code: '' })).toBe('Write tests');
  });

  it('maps a task into form values', () => {
    expect(getTaskFormValues(createTask())).toEqual({
      project_column_id: 'column-1',
      code: 'FRONTEND-3',
      title: 'Implement dependencies',
      description: '<p>Details</p>',
      due_date: '2026-06-15',
      priority: 'medium',
      responsible_id: null,
      tags: 'frontend,forms',
      depends_on_task_ids: ['task-2', 'task-3'],
    });
  });
});

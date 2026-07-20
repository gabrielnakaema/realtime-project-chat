import { describe, expect, it } from 'vitest';
import { getNextArchivedTasksPageParam } from './use-archived-tasks-list';
import type { CursorPaginated } from '@/shared/types/paginated';
import type { Task } from '@/features/tasks/types/task';

const createTask = (id: string, order: string, updatedAt: string): Task => ({
  id,
  project_id: 'project-1',
  title: `Task ${id}`,
  description: '',
  code: '',
  status: 'todo',
  project_column_id: 'column-1',
  project_column: null,
  created_at: updatedAt,
  updated_at: updatedAt,
  priority: 'medium',
  order,
  version: 0,
  responsible_id: null,
  due_date: null,
  done_at: null,
  archived_at: updatedAt,
  tags: null,
  author_id: 'user-1',
  author: {
    id: 'user-1',
    name: 'Test User',
    email: 'user@example.com',
  },
  responsible: null,
  updates: [],
  project: null,
});

const createPage = (tasks: Task[], hasNext: boolean): CursorPaginated<Task> => ({
  data: tasks,
  has_next: hasNext,
  has_previous: false,
});

describe('getNextArchivedTasksPageParam', () => {
  it('returns the last task cursor when another page exists', () => {
    const lastPage = createPage(
      [
        createTask('task-1', '001', '2026-06-01T10:00:00.000Z'),
        createTask('task-2', '002', '2026-06-02T10:00:00.000Z'),
      ],
      true,
    );

    expect(getNextArchivedTasksPageParam(lastPage)).toEqual({
      taskOrder: '002',
      updatedAt: '2026-06-02T10:00:00.000Z',
    });
  });

  it('returns undefined when there is no next page', () => {
    const lastPage = createPage([createTask('task-1', '001', '2026-06-01T10:00:00.000Z')], false);

    expect(getNextArchivedTasksPageParam(lastPage)).toBeUndefined();
  });

  it('returns undefined for an empty page', () => {
    expect(getNextArchivedTasksPageParam(createPage([], true))).toBeUndefined();
  });
});

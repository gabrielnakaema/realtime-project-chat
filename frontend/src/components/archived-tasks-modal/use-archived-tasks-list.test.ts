import { describe, expect, it } from 'vitest';
import {
  buildArchivedTasksRequest,
  flattenArchivedTasks,
  getNextArchivedTasksPageParam,
} from './use-archived-tasks-list';
import type { CursorPaginated } from '@/types/paginated';
import type { Task } from '@/types/task';
import { DEFAULT_TASK_LIMIT } from '@/constants/tasks';

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

describe('buildArchivedTasksRequest', () => {
  it('builds the archived tasks request with the default page size', () => {
    expect(buildArchivedTasksRequest('project-1')).toEqual({
      projectId: 'project-1',
      projectColumnIds: [],
      archived: true,
      taskOrder: '',
      updatedAt: null,
      limit: DEFAULT_TASK_LIMIT,
    });
  });
});

describe('getNextArchivedTasksPageParam', () => {
  it('returns the last task cursor when another page exists', () => {
    const lastPage = {
      'column-1': createPage(
        [
          createTask('task-1', '001', '2026-06-01T10:00:00.000Z'),
          createTask('task-2', '002', '2026-06-02T10:00:00.000Z'),
        ],
        true,
      ),
    };

    expect(getNextArchivedTasksPageParam(lastPage)).toEqual({
      taskOrder: '002',
      updatedAt: '2026-06-02T10:00:00.000Z',
    });
  });

  it('returns undefined when there is no next page', () => {
    const lastPage = {
      'column-1': createPage([createTask('task-1', '001', '2026-06-01T10:00:00.000Z')], false),
    };

    expect(getNextArchivedTasksPageParam(lastPage)).toBeUndefined();
  });
});

describe('flattenArchivedTasks', () => {
  it('collects tasks from every column and page', () => {
    const pages: NonNullable<Parameters<typeof flattenArchivedTasks>[0]> = [
      {
        'column-1': createPage([createTask('task-1', '001', '2026-06-01T10:00:00.000Z')], true),
        'column-2': createPage([createTask('task-2', '002', '2026-06-02T10:00:00.000Z')], true),
      },
      {
        'column-1': createPage([createTask('task-3', '003', '2026-06-03T10:00:00.000Z')], false),
      },
    ];

    expect(flattenArchivedTasks(pages).map((task) => task.id)).toEqual(['task-1', 'task-2', 'task-3']);
  });
});

import { describe, expect, it } from 'vitest';
import { getNotificationNavigationTarget } from './notification-navigation';
import type { Notification } from '@/types/notification';

const baseNotification: Notification = {
  id: 'notification-1',
  user_id: 'user-1',
  actor_id: 'user-2',
  project_id: 'project-fallback',
  task_id: null,
  task_comment_id: null,
  type: 'task.comment.created',
  read_at: null,
  created_at: '2026-05-25T10:00:00.000Z',
  updated_at: '2026-05-25T10:00:00.000Z',
  actor: null,
  project: null,
  task: null,
  task_comment: null,
};

describe('getNotificationNavigationTarget', () => {
  it('maps nested task and comment deep-link data into route search params', () => {
    expect(
      getNotificationNavigationTarget({
        ...baseNotification,
        project: {
          id: 'project-1',
          user_id: 'user-1',
          name: 'Project',
          description: '',
          repository_url: '',
          repository_owner: '',
          repository_name: '',
          default_branch: '',
          branch_name_prefix: '',
          created_at: '',
          updated_at: '',
          columns: [],
          members: [],
        },
        task: {
          id: 'task-1',
          project_id: 'project-1',
          title: 'Task',
          description: '',
          code: 'TASK-1',
          status: 'todo',
          project_column_id: 'column-1',
          project_column: null,
          created_at: '',
          updated_at: '',
          priority: 'medium',
          order: 'a0',
          version: 0,
          responsible_id: null,
          due_date: null,
          done_at: null,
          archived_at: null,
          tags: [],
          author_id: 'user-1',
          author: { id: 'user-1', name: 'User', email: 'user@example.com' },
          responsible: null,
          updates: [],
          project: null,
        },
        task_comment: {
          id: 'comment-1',
          task: null,
          user: { id: 'user-1', name: 'User', email: 'user@example.com' },
          content: '<p>Comment</p>',
          created_at: '2026-05-25T12:00:00.000Z',
          updated_at: '2026-05-25T12:00:00.000Z',
          replies: [],
        },
      }),
    ).toEqual({
      to: '/projects/$projectId',
      params: { projectId: 'project-1' },
      search: {
        taskId: 'task-1',
        commentId: 'comment-1',
        commentCreatedAt: '2026-05-25T12:00:00.000Z',
      },
    });
  });

  it('falls back to raw ids when embedded entities are missing', () => {
    expect(
      getNotificationNavigationTarget({
        ...baseNotification,
        project_id: 'project-2',
        task_id: 'task-2',
        task_comment_id: 'comment-2',
      }),
    ).toEqual({
      to: '/projects/$projectId',
      params: { projectId: 'project-2' },
      search: {
        taskId: 'task-2',
        commentId: 'comment-2',
        commentCreatedAt: undefined,
      },
    });
  });
});

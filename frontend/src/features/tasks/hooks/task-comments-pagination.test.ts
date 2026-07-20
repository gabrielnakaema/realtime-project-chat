import { describe, expect, it } from 'vitest';
import {
  TASK_COMMENTS_PAGE_SIZE,
  buildTaskCommentsQueryKey,
  buildTaskCommentsRequest,
  getInitialCommentsCursor,
  getNextCommentsPageParam,
  getPreviousCommentsPageParam,
} from './task-comments-pagination';
import type { CommentsPageParam } from './task-comments-pagination';
import type { CursorPaginated } from '@/shared/types/paginated';
import type { TaskComment } from '@/features/tasks/types/task';

const createComment = (id: string, createdAt: string): TaskComment => ({
  id,
  task: null,
  user: {
    id: 'user-1',
    name: 'Test User',
    email: 'user@example.com',
  },
  content: '<p>Comment</p>',
  action_origin: 'user',
  created_at: createdAt,
  updated_at: createdAt,
  replies: [],
});

describe('buildTaskCommentsQueryKey', () => {
  it('includes task anchor params and page size in the key', () => {
    expect(
      buildTaskCommentsQueryKey({
        taskId: 'task-1',
        targetCommentId: 'comment-1',
        targetCommentCreatedAt: '2026-05-25T10:00:00.000Z',
      }),
    ).toEqual([
      'tasks',
      'comments',
      {
        taskId: 'task-1',
        commentId: 'comment-1',
        commentCreatedAt: '2026-05-25T10:00:00.000Z',
        limit: TASK_COMMENTS_PAGE_SIZE,
      },
    ]);
  });
});

describe('buildTaskCommentsRequest', () => {
  it('builds the initial request from the anchor comment', () => {
    expect(
      buildTaskCommentsRequest({
        taskId: 'task-1',
        pageParam: { direction: 'initial' },
        targetCommentId: 'comment-1',
        targetCommentCreatedAt: '2026-05-25T10:00:00.000Z',
      }),
    ).toEqual({
      taskId: 'task-1',
      limit: TASK_COMMENTS_PAGE_SIZE,
      before: '2026-05-25T10:00:00.001Z',
      beforeCommentId: 'comment-1',
    });
  });

  it('builds a before-cursor request', () => {
    const pageParam: CommentsPageParam = {
      direction: 'before',
      cursor: '2026-05-24T10:00:00.000Z',
      commentId: 'comment-before',
    };

    expect(buildTaskCommentsRequest({ taskId: 'task-1', pageParam })).toEqual({
      taskId: 'task-1',
      limit: TASK_COMMENTS_PAGE_SIZE,
      before: '2026-05-24T10:00:00.000Z',
      beforeCommentId: 'comment-before',
    });
  });

  it('builds an after-cursor request', () => {
    const pageParam: CommentsPageParam = {
      direction: 'after',
      cursor: '2026-05-26T10:00:00.000Z',
      commentId: 'comment-after',
    };

    expect(buildTaskCommentsRequest({ taskId: 'task-1', pageParam })).toEqual({
      taskId: 'task-1',
      limit: TASK_COMMENTS_PAGE_SIZE,
      after: '2026-05-26T10:00:00.000Z',
      afterCommentId: 'comment-after',
    });
  });
});

describe('cursor helpers', () => {
  it('returns undefined for invalid initial cursor dates', () => {
    expect(getInitialCommentsCursor('not-a-date')).toBeUndefined();
  });

  it('derives the next page param from the last comment', () => {
    const page: CursorPaginated<TaskComment> = {
      data: [
        createComment('comment-1', '2026-05-24T10:00:00.000Z'),
        createComment('comment-2', '2026-05-23T10:00:00.000Z'),
      ],
      has_next: true,
      has_previous: false,
    };

    expect(getNextCommentsPageParam(page)).toEqual({
      direction: 'before',
      cursor: '2026-05-23T10:00:00.000Z',
      commentId: 'comment-2',
    });
  });

  it('derives the previous page param from the first comment', () => {
    const page: CursorPaginated<TaskComment> = {
      data: [
        createComment('comment-1', '2026-05-26T10:00:00.000Z'),
        createComment('comment-2', '2026-05-25T10:00:00.000Z'),
      ],
      has_next: false,
      has_previous: true,
    };

    expect(getPreviousCommentsPageParam(page)).toEqual({
      direction: 'after',
      cursor: '2026-05-26T10:00:00.000Z',
      commentId: 'comment-1',
    });
  });

  it('returns undefined when pagination is exhausted', () => {
    const page: CursorPaginated<TaskComment> = {
      data: [createComment('comment-1', '2026-05-26T10:00:00.000Z')],
      has_next: false,
      has_previous: false,
    };

    expect(getNextCommentsPageParam(page)).toBeUndefined();
    expect(getPreviousCommentsPageParam(page)).toBeUndefined();
  });
});

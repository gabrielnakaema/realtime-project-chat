import type { CursorPaginated } from '@/shared/types/paginated';
import type { TaskComment } from '@/features/tasks/types/task';
import type { TaskCommentsQueryKey } from '@/shared/services/query-keys';
import { taskQueryKeys } from '@/shared/services/query-keys';

export type CommentsPageParam =
  | { direction: 'initial' }
  | { direction: 'before'; cursor: string; commentId: string }
  | { direction: 'after'; cursor: string; commentId: string };

export interface ListTaskCommentsPageRequest {
  taskId: string;
  limit: number;
  before?: string;
  beforeCommentId?: string;
  after?: string;
  afterCommentId?: string;
}

export const TASK_COMMENTS_PAGE_SIZE = 20;
const INITIAL_COMMENT_PAGE_OFFSET_MS = 1;

export const buildTaskCommentsQueryKey = ({
  taskId,
  targetCommentId,
  targetCommentCreatedAt,
  limit = TASK_COMMENTS_PAGE_SIZE,
}: {
  taskId: string;
  targetCommentId?: string;
  targetCommentCreatedAt?: string;
  limit?: number;
}): TaskCommentsQueryKey =>
  taskQueryKeys.comments(taskId, {
    commentId: targetCommentId,
    commentCreatedAt: targetCommentCreatedAt,
    limit,
  });

export const buildTaskCommentsRequest = ({
  taskId,
  pageParam,
  targetCommentId,
  targetCommentCreatedAt,
  limit = TASK_COMMENTS_PAGE_SIZE,
}: {
  taskId: string;
  pageParam: CommentsPageParam;
  targetCommentId?: string;
  targetCommentCreatedAt?: string;
  limit?: number;
}): ListTaskCommentsPageRequest => {
  if (pageParam.direction === 'before') {
    return {
      taskId,
      limit,
      before: pageParam.cursor,
      beforeCommentId: pageParam.commentId,
    };
  }

  if (pageParam.direction === 'after') {
    return {
      taskId,
      limit,
      after: pageParam.cursor,
      afterCommentId: pageParam.commentId,
    };
  }

  return {
    taskId,
    limit,
    before: getInitialCommentsCursor(targetCommentCreatedAt),
    beforeCommentId: targetCommentId,
  };
};

export const getNextCommentsPageParam = (lastPage: CursorPaginated<TaskComment>): CommentsPageParam | undefined => {
  if (!lastPage.has_next || lastPage.data.length === 0) {
    return undefined;
  }

  const lastComment = lastPage.data[lastPage.data.length - 1];
  return {
    direction: 'before',
    cursor: lastComment.created_at,
    commentId: lastComment.id,
  };
};

export const getPreviousCommentsPageParam = (
  firstPage: CursorPaginated<TaskComment>,
): CommentsPageParam | undefined => {
  if (!firstPage.has_previous || firstPage.data.length === 0) {
    return undefined;
  }

  const firstComment = firstPage.data[0];
  return {
    direction: 'after',
    cursor: firstComment.created_at,
    commentId: firstComment.id,
  };
};

export const getInitialCommentsCursor = (commentCreatedAt?: string) => {
  if (!commentCreatedAt) {
    return undefined;
  }

  const createdAt = new Date(commentCreatedAt);
  if (Number.isNaN(createdAt.getTime())) {
    return undefined;
  }

  return new Date(createdAt.getTime() + INITIAL_COMMENT_PAGE_OFFSET_MS).toISOString();
};

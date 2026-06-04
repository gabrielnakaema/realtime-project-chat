import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent } from 'react';
import type { InfiniteData } from '@tanstack/react-query';
import type { RefCallback } from 'react';
import type { CursorPaginated } from '@/types/paginated';
import type { TaskComment } from '@/types/task';
import type { SocketEvent } from '@/types/websocket';
import type { CommentsPageParam } from '@/hooks/task-comments-pagination';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import {
  TASK_COMMENTS_PAGE_SIZE,
  buildTaskCommentsQueryKey,
  buildTaskCommentsRequest,
  getNextCommentsPageParam,
  getPreviousCommentsPageParam,
} from '@/hooks/task-comments-pagination';
import { useSocket } from '@/hooks/use-socket';
import { createTaskComment, listTaskComments } from '@/services/tasks';

type CommentsInfiniteData = InfiniteData<CursorPaginated<TaskComment>, CommentsPageParam>;

export interface UseTaskCommentsParams {
  taskId: string;
  projectId?: string;
  open: boolean;
  targetCommentId?: string;
  targetCommentCreatedAt?: string;
}

export interface SubmitTaskCommentInput {
  content: string;
  parentCommentId?: string | null;
}

export type SubmitTaskComment = (input: SubmitTaskCommentInput) => Promise<TaskComment>;

export interface UseTaskCommentsResult {
  comments: TaskComment[];
  isLoading: boolean;
  isSubmitting: boolean;
  hasPreviousPage: boolean;
  isFetchingNextPage: boolean;
  isFetchingPreviousPage: boolean;
  fetchPreviousPage: () => Promise<unknown>;
  sentinelRef: RefCallback<HTMLDivElement>;
  submitComment: SubmitTaskComment;
}

export const useTaskComments = ({
  taskId,
  projectId,
  open,
  targetCommentId,
  targetCommentCreatedAt,
}: UseTaskCommentsParams): UseTaskCommentsResult => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();

  const commentsQueryKey = buildTaskCommentsQueryKey({
    taskId,
    targetCommentId,
    targetCommentCreatedAt,
  });

  const {
    data,
    isLoading,
    fetchNextPage,
    fetchPreviousPage,
    hasPreviousPage,
    hasNextPage,
    isFetchingNextPage,
    isFetchingPreviousPage,
  } = useInfiniteQuery({
    queryKey: commentsQueryKey,
    queryFn: ({ pageParam, queryKey }) => {
      const [, , key] = queryKey;

      return listTaskComments(
        buildTaskCommentsRequest({
          taskId: key.taskId,
          pageParam,
          targetCommentId: key.commentId ?? undefined,
          targetCommentCreatedAt: key.commentCreatedAt ?? undefined,
          limit: key.limit ?? TASK_COMMENTS_PAGE_SIZE,
        }),
      );
    },
    getNextPageParam: getNextCommentsPageParam,
    getPreviousPageParam: getPreviousCommentsPageParam,
    initialPageParam: { direction: 'initial' } as CommentsPageParam,
    enabled: open,
  });

  const comments = data?.pages.flatMap((page) => page.data) ?? [];

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: () => {
      if (!hasNextPage || isFetchingNextPage) {
        return;
      }

      fetchNextPage();
    },
  });

  const { mutateAsync: submitComment, isPending: isSubmitting } = useMutation({
    mutationFn: ({ content, parentCommentId }: SubmitTaskCommentInput) =>
      createTaskComment({ taskId, content, parentCommentId }),
    onSuccess: (created) => {
      queryClient.setQueryData<CommentsInfiniteData>(commentsQueryKey, (current) =>
        prependCommentToInfiniteData(current, created),
      );
    },
  });

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type !== 'task_comment_created') return;
    if (event.data.task?.id !== taskId) return;

    queryClient.setQueryData<CommentsInfiniteData>(commentsQueryKey, (current) =>
      prependCommentToInfiniteData(current, event.data),
    );
  });

  useEffect(() => {
    if (!open || !projectId || status !== 'connected') return;
    const unsubscribe = subscribe(projectId, 'project', handleSocketEvent);
    return () => unsubscribe();
  }, [open, status, subscribe, projectId]);

  return {
    comments,
    isLoading,
    isSubmitting,
    hasPreviousPage: Boolean(hasPreviousPage),
    isFetchingNextPage,
    isFetchingPreviousPage,
    fetchPreviousPage: async () => fetchPreviousPage(),
    sentinelRef,
    submitComment,
  };
};

export const prependCommentToInfiniteData = (
  current: CommentsInfiniteData | undefined,
  incoming: TaskComment,
): CommentsInfiniteData => {
  const baseData: CommentsInfiniteData = current ?? {
    pages: [{ data: [], has_next: false, has_previous: false }],
    pageParams: [{ direction: 'initial' }],
  };

  const allComments = baseData.pages.flatMap((page) => page.data);

  if (commentExists(allComments, incoming.id)) {
    return baseData;
  }

  if (incoming.parent_comment_id) {
    const parentCommentId = incoming.parent_comment_id;

    const parentLoaded = allComments.some(
      (c) => c.id === parentCommentId || commentExists(c.replies, parentCommentId),
    );

    if (!parentLoaded) {
      return baseData;
    }

    return {
      ...baseData,
      pages: baseData.pages.map((page) => ({
        ...page,
        data: insertReply(page.data, parentCommentId, incoming),
      })),
    };
  }

  const [firstPage, ...restPages] = baseData.pages;
  return {
    ...baseData,
    pages: [{ ...firstPage, data: [incoming, ...firstPage.data] }, ...restPages],
  };
};

const commentExists = (comments: TaskComment[], id: string): boolean => {
  return comments.some((comment) => comment.id === id || commentExists(comment.replies, id));
};

const insertReply = (comments: TaskComment[], parentId: string, incoming: TaskComment): TaskComment[] => {
  return comments.map((comment) => {
    if (comment.id === parentId) {
      return {
        ...comment,
        replies: commentExists(comment.replies, incoming.id) ? comment.replies : [incoming, ...comment.replies],
      };
    }

    if (comment.replies.length === 0) {
      return comment;
    }

    return { ...comment, replies: insertReply(comment.replies, parentId, incoming) };
  });
};

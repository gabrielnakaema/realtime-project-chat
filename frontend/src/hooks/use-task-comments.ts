import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useState } from 'react';
import type { InfiniteData } from '@tanstack/react-query';
import type { CursorPaginated } from '@/types/paginated';
import type { TaskComment } from '@/types/task';
import type { SocketEvent } from '@/types/websocket';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { useSocket } from '@/hooks/use-socket';
import { createTaskComment, listTaskComments } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';

type CommentsInfiniteData = InfiniteData<CursorPaginated<TaskComment>, string | null>;

export const useTaskComments = ({ taskId, projectId, open }: { taskId: string; projectId?: string; open: boolean }) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();
  const [commentDraft, setCommentDraft] = useState('');
  const [replyDraft, setReplyDraft] = useState('');
  const [replyingToId, setReplyingToId] = useState<string | null>(null);
  const [composerKey, setComposerKey] = useState(0);
  const [replyEditorKey, setReplyEditorKey] = useState(0);

  const queryKey = taskQueryKeys.comments(taskId);

  const { data, isLoading, fetchNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: taskQueryKeys.comments(taskId),
    queryFn: ({ pageParam }) => listTaskComments({ taskId, limit: 20, before: pageParam ?? undefined }),
    getNextPageParam: (lastPage) => {
      if (!lastPage.has_next) return undefined;
      return lastPage.data[lastPage.data.length - 1]?.created_at ?? undefined;
    },
    initialPageParam: null as string | null,
    enabled: open,
  });

  const comments = data?.pages.flatMap((page) => page.data) ?? [];

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
  });

  const { mutate: submitMutation, isPending: isSubmitting } = useMutation({
    mutationFn: ({ content, parentCommentId }: { content: string; parentCommentId?: string | null }) =>
      createTaskComment({ taskId, content, parentCommentId }),
    onSuccess: (created, variables) => {
      queryClient.setQueryData<CommentsInfiniteData>(queryKey, (current) =>
        prependCommentToInfiniteData(current, created),
      );

      if (variables.parentCommentId) {
        setReplyDraft('');
        setReplyingToId(null);
        setReplyEditorKey((k) => k + 1);
      } else {
        setCommentDraft('');
        setComposerKey((k) => k + 1);
      }
    },
  });

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type !== 'task_comment_created') return;
    if (event.data.task?.id !== taskId) return;

    queryClient.setQueryData<CommentsInfiniteData>(queryKey, (current) =>
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
    isFetchingNextPage,
    sentinelRef,
    commentDraft,
    setCommentDraft,
    composerKey,
    replyDraft,
    setReplyDraft,
    replyEditorKey,
    replyingToId,
    submitComment: () => submitMutation({ content: commentDraft, parentCommentId: null }),
    startReply: (commentId: string) => {
      setReplyingToId((current) => (current === commentId ? null : commentId));
      setReplyDraft('');
    },
    cancelReply: () => {
      setReplyingToId(null);
      setReplyDraft('');
    },
    submitReply: (parentCommentId: string) => submitMutation({ content: replyDraft, parentCommentId }),
  };
};

const prependCommentToInfiniteData = (
  current: CommentsInfiniteData | undefined,
  incoming: TaskComment,
): CommentsInfiniteData => {
  const baseData: CommentsInfiniteData = current ?? {
    pages: [{ data: [], has_next: false }],
    pageParams: [null],
  };

  const allComments = baseData.pages.flatMap((page) => page.data);

  if (commentExists(allComments, incoming.id)) {
    return baseData;
  }

  if (incoming.parent_comment_id) {
    return {
      ...baseData,
      pages: baseData.pages.map((page) => ({
        ...page,
        data: insertReply(page.data, incoming.parent_comment_id!, incoming),
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

    if (comment.replies.length === 0) return comment;

    return { ...comment, replies: insertReply(comment.replies, parentId, incoming) };
  });
};

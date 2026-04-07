import { MessageSquare, Reply, SendHorizontal } from 'lucide-react';
import { Button } from './button';
import { LoadingSpinner } from './loading';
import { TextEditor } from './text-editor';
import { ScrollArea, ScrollBar } from './ui/scroll-area';
import type { TaskComment } from '@/types/task';
import { useAuth } from '@/hooks/use-auth';
import { useTaskComments } from '@/hooks/use-task-comments';
import { cn } from '@/lib/utils';
import { formatDateString } from '@/utils/date';
import { isHtmlContentEmpty, sanitizeHTML } from '@/utils/html';

interface TaskCommentsProps {
  taskId: string;
  projectId?: string;
  open: boolean;
}

export const TaskComments = ({ taskId, projectId, open }: TaskCommentsProps) => {
  const { user } = useAuth();
  const {
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
    submitComment,
    startReply,
    cancelReply,
    submitReply,
  } = useTaskComments({ taskId, projectId, open });

  const renderComments = () => {
    if (isLoading) {
      return (
        <div className="flex min-h-40 items-center justify-center">
          <LoadingSpinner size="2rem" />
        </div>
      );
    }

    if (comments.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-300 px-6 py-12 text-center dark:border-slate-700">
          <MessageSquare className="mb-3 h-8 w-8 text-slate-400 dark:text-slate-500" />
          <p className="text-sm font-medium text-slate-700 dark:text-slate-200">No comments yet</p>
          <p className="mt-1 max-w-sm text-sm text-slate-500 dark:text-slate-400">
            Use comments to capture decisions, ask for feedback, and keep discussion attached to the task.
          </p>
        </div>
      );
    }

    return (
      <>
        {comments.map((comment) => (
          <TaskCommentItem
            key={comment.id}
            comment={comment}
            currentUserId={user?.id}
            replyingToId={replyingToId}
            replyDraft={replyDraft}
            onReplyDraftChange={setReplyDraft}
            onReplyStart={startReply}
            onReplyCancel={cancelReply}
            onReplySubmit={submitReply}
            isSubmittingReply={isSubmitting}
            replyEditorKey={replyEditorKey}
          />
        ))}
        {isFetchingNextPage && (
          <div className="flex justify-center py-2">
            <LoadingSpinner size="1.25rem" />
          </div>
        )}
        <div ref={sentinelRef} className="h-1" aria-hidden="true" />
      </>
    );
  };

  return (
    <div className="border-t border-slate-200 bg-slate-50/40 xl:min-h-0 xl:overflow-hidden xl:border-t-0 xl:border-l dark:border-slate-700 dark:bg-slate-950/20">
      <div className="flex flex-col xl:h-full xl:min-h-0">
        <div className="border-b border-slate-200 px-6 py-5 dark:border-slate-700">
          <div className="mb-4 flex items-center gap-3">
            <div className="rounded-full bg-slate-100 p-2 dark:bg-slate-800">
              <MessageSquare className="h-4 w-4 text-slate-600 dark:text-slate-300" />
            </div>
            <div>
              <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">Comments</p>
            </div>
          </div>

          <TextEditor
            key={composerKey}
            initialValue=""
            onChange={setCommentDraft}
            placeholder="Share context, decisions, or blockers..."
          />
          <div className="mt-3 flex items-center justify-between gap-3">
            <Button
              type="button"
              onClick={submitComment}
              disabled={isSubmitting || isHtmlContentEmpty(commentDraft)}
              className="ml-auto shrink-0"
            >
              {isSubmitting ? <LoadingSpinner size="1rem" /> : <SendHorizontal className="h-4 w-4" />}
              Comment
            </Button>
          </div>
        </div>

        <ScrollArea className="xl:min-h-0 xl:flex-1">
          <div className="min-w-max px-6 py-5">
            <div className="flex flex-col gap-4">{renderComments()}</div>
          </div>
          <ScrollBar orientation="horizontal" />
        </ScrollArea>
      </div>
    </div>
  );
};

interface TaskCommentItemProps {
  comment: TaskComment;
  currentUserId?: string;
  replyingToId: string | null;
  replyDraft: string;
  onReplyDraftChange: (value: string) => void;
  onReplyStart: (commentId: string) => void;
  onReplyCancel: () => void;
  onReplySubmit: (parentCommentId: string) => void;
  isSubmittingReply: boolean;
  replyEditorKey: number;
  depth?: number;
}

const TaskCommentItem = ({
  comment,
  currentUserId,
  replyingToId,
  replyDraft,
  onReplyDraftChange,
  onReplyStart,
  onReplyCancel,
  onReplySubmit,
  isSubmittingReply,
  replyEditorKey,
  depth = 0,
}: TaskCommentItemProps) => {
  const isReplying = replyingToId === comment.id;
  const isOwnComment = currentUserId === comment.user.id;

  return (
    <div className={cn('space-y-3', depth > 0 && 'ml-4 border-l border-slate-200 pl-4 dark:border-slate-700')}>
      <article className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-700 dark:bg-slate-900">
        <div className="mb-3 flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <p className="truncate text-sm font-semibold text-slate-900 dark:text-slate-100">{comment.user.name}</p>
              {isOwnComment && (
                <span className="rounded-full bg-blue-50 px-2 py-0.5 text-[11px] font-medium text-blue-700 dark:bg-blue-950/60 dark:text-blue-300">
                  You
                </span>
              )}
            </div>
            <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">{formatDateString(comment.created_at)}</p>
          </div>

          <button
            type="button"
            className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200"
            onClick={() => onReplyStart(comment.id)}
          >
            <Reply className="h-3.5 w-3.5" />
            Reply
          </button>
        </div>

        <div
          className="prose prose-sm prose-slate dark:prose-invert max-w-none"
          dangerouslySetInnerHTML={{ __html: sanitizeHTML(comment.content) }}
        />

        {isReplying && (
          <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-950/30">
            <TextEditor
              key={`${comment.id}-${replyEditorKey}`}
              initialValue=""
              onChange={onReplyDraftChange}
              placeholder={`Reply to ${comment.user.name}...`}
            />
            <div className="mt-3 flex items-center justify-end gap-2">
              <Button type="button" variant="secondary" onClick={onReplyCancel}>
                Cancel
              </Button>
              <Button
                type="button"
                onClick={() => onReplySubmit(comment.id)}
                disabled={isSubmittingReply || isHtmlContentEmpty(replyDraft)}
              >
                {isSubmittingReply ? <LoadingSpinner size="1rem" /> : 'Reply'}
              </Button>
            </div>
          </div>
        )}
      </article>

      {comment.replies.length > 0 && (
        <div className="space-y-3">
          {comment.replies.map((reply) => (
            <TaskCommentItem
              key={reply.id}
              comment={reply}
              currentUserId={currentUserId}
              replyingToId={replyingToId}
              replyDraft={replyDraft}
              onReplyDraftChange={onReplyDraftChange}
              onReplyStart={onReplyStart}
              onReplyCancel={onReplyCancel}
              onReplySubmit={onReplySubmit}
              isSubmittingReply={isSubmittingReply}
              replyEditorKey={replyEditorKey}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
};

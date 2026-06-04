import { Reply } from 'lucide-react';
import { useState } from 'react';
import { ActionOriginBadge } from '../action-origin-badge';
import { ReplyComposer } from './reply-composer';
import type { SubmitTaskComment } from './use-task-comments';
import type { TaskComment } from '@/types/task';
import { useAuth } from '@/hooks/use-auth';
import { cn } from '@/lib/utils';
import { formatDateString } from '@/utils/date';
import { isHtmlContentEmpty, sanitizeHTML } from '@/utils/html';

interface TaskCommentItemProps {
  comment: TaskComment;
  isSubmittingReply: boolean;
  onSubmitReply: SubmitTaskComment;
  depth?: number;
}

export const TaskCommentItem = ({ comment, isSubmittingReply, onSubmitReply, depth = 0 }: TaskCommentItemProps) => {
  const { user } = useAuth();
  const [isReplyComposerOpen, setIsReplyComposerOpen] = useState(false);
  const [replyDraft, setReplyDraft] = useState('');
  const [replyComposerKey, setReplyComposerKey] = useState(0);
  const isOwnComment = user?.id === comment.user.id;

  const resetReplyComposer = () => {
    setIsReplyComposerOpen(false);
    setReplyDraft('');
    setReplyComposerKey((current) => current + 1);
  };

  const handleSubmitReply = async () => {
    if (isHtmlContentEmpty(replyDraft)) {
      return;
    }

    await onSubmitReply({ content: replyDraft, parentCommentId: comment.id });
    resetReplyComposer();
  };

  return (
    <div className={cn('min-w-0 space-y-3', depth > 0 && 'ml-4 border-l border-slate-200 pl-4 dark:border-slate-700')}>
      <article
        id={`task-comment-${comment.id}`}
        className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition-[background-color,transform] duration-700 ease-out dark:border-slate-700 dark:bg-slate-900"
      >
        <div className="mb-3 flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <p className="truncate text-sm font-semibold text-slate-900 dark:text-slate-100">{comment.user.name}</p>
              <ActionOriginBadge origin={comment.action_origin} />
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
            onClick={() => setIsReplyComposerOpen(true)}
          >
            <Reply className="h-3.5 w-3.5" />
            Reply
          </button>
        </div>

        <div
          className="prose prose-sm prose-slate dark:prose-invert max-w-none text-[0.95rem] leading-7 text-pretty [overflow-wrap:anywhere] [&_*]:max-w-full [&_a]:break-words [&_code]:whitespace-pre-wrap [&_li]:marker:text-slate-400 [&_ol]:pl-5 [&_p]:my-3 [&_pre]:overflow-x-auto [&_pre]:break-words [&_pre]:whitespace-pre-wrap [&_ul]:pl-5"
          dangerouslySetInnerHTML={{ __html: sanitizeHTML(comment.content) }}
        />

        {isReplyComposerOpen && (
          <ReplyComposer
            comment={comment}
            composerKey={replyComposerKey}
            draft={replyDraft}
            isSubmitting={isSubmittingReply}
            onCancel={resetReplyComposer}
            onDraftChange={setReplyDraft}
            onSubmit={handleSubmitReply}
          />
        )}
      </article>

      {comment.replies.length > 0 && (
        <div className="space-y-3">
          {comment.replies.map((reply) => (
            <TaskCommentItem
              key={reply.id}
              comment={reply}
              isSubmittingReply={isSubmittingReply}
              onSubmitReply={onSubmitReply}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
};

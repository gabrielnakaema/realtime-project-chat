import { EmptyTaskCommentsState } from './empty-task-comments-state';
import { TaskCommentItem } from './task-comment-item';
import type { UseTaskCommentsResult } from './use-task-comments';
import { LoadingSpinner } from '@/shared/components/loading';
import { Button } from '@/shared/components/button';

interface TaskCommentsFeedProps {
  taskComments: UseTaskCommentsResult;
}

export const TaskCommentsFeed = ({ taskComments }: TaskCommentsFeedProps) => {
  const {
    comments,
    fetchPreviousPage,
    hasPreviousPage,
    isFetchingNextPage,
    isFetchingPreviousPage,
    isLoading,
    isSubmitting,
    sentinelRef,
    submitComment,
  } = taskComments;

  if (isLoading) {
    return (
      <div className="flex min-h-40 items-center justify-center">
        <LoadingSpinner size="2rem" />
      </div>
    );
  }

  if (comments.length === 0) {
    return <EmptyTaskCommentsState />;
  }

  return (
    <div className="flex flex-col gap-4">
      {(hasPreviousPage || isFetchingPreviousPage) && (
        <div className="flex justify-center pb-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => fetchPreviousPage()}
            disabled={isFetchingPreviousPage}
            className="text-muted-foreground hover:bg-card bg-transparent"
          >
            {isFetchingPreviousPage ? <LoadingSpinner size="1rem" /> : 'Load recent comments'}
          </Button>
        </div>
      )}

      {comments.map((comment) => (
        <TaskCommentItem
          key={comment.id}
          comment={comment}
          isSubmittingReply={isSubmitting}
          onSubmitReply={submitComment}
        />
      ))}

      {isFetchingNextPage && (
        <div className="flex justify-center py-2">
          <LoadingSpinner size="1.25rem" />
        </div>
      )}

      <div ref={sentinelRef} className="h-1" aria-hidden="true" />
    </div>
  );
};

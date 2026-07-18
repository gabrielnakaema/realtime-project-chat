import { MessageSquare } from 'lucide-react';
import { ScrollArea } from '../ui/scroll-area';
import { TaskCommentComposer } from './task-comment-composer';
import { TaskCommentsFeed } from './task-comments-feed';
import { useTargetCommentHighlight } from './use-target-comment-highlight';
import { useTaskComments } from './use-task-comments';

interface TaskCommentsProps {
  taskId: string;
  projectId?: string;
  open: boolean;
  targetCommentId?: string;
  targetCommentCreatedAt?: string;
}

export const TaskComments = ({
  taskId,
  projectId,
  open,
  targetCommentId,
  targetCommentCreatedAt,
}: TaskCommentsProps) => {
  const taskComments = useTaskComments({ taskId, projectId, open, targetCommentId, targetCommentCreatedAt });

  useTargetCommentHighlight({ open, targetCommentId });

  return (
    <div className="border-border bg-muted/40 border-t xl:h-full xl:min-h-0 xl:overflow-hidden xl:border-t-0 xl:border-l">
      <ScrollArea className="xl:h-full xl:min-h-0">
        <div className="flex min-h-full flex-col">
          <div className="border-border border-b px-6 py-5">
            <div className="mb-4 flex items-center gap-3">
              <div className="bg-muted rounded-full p-2">
                <MessageSquare className="text-muted-foreground h-4 w-4" />
              </div>
              <div>
                <p className="text-foreground text-sm font-semibold">Comments</p>
              </div>
            </div>

            <TaskCommentComposer isSubmitting={taskComments.isSubmitting} onSubmit={taskComments.submitComment} />
          </div>

          <div className="min-w-0 px-6 py-5">
            <TaskCommentsFeed taskComments={taskComments} />
          </div>
        </div>
      </ScrollArea>
    </div>
  );
};

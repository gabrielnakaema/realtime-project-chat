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
    <div className="border-t border-slate-200 bg-slate-50/40 xl:h-full xl:min-h-0 xl:overflow-hidden xl:border-t-0 xl:border-l dark:border-slate-700 dark:bg-slate-950/20">
      <ScrollArea className="xl:h-full xl:min-h-0">
        <div className="flex min-h-full flex-col">
          <div className="border-b border-slate-200 px-6 py-5 dark:border-slate-700">
            <div className="mb-4 flex items-center gap-3">
              <div className="rounded-full bg-slate-100 p-2 dark:bg-slate-800">
                <MessageSquare className="h-4 w-4 text-slate-600 dark:text-slate-300" />
              </div>
              <div>
                <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">Comments</p>
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

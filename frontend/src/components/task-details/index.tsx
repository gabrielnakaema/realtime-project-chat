import { X } from 'lucide-react';
import { LoadingSpinner } from '../loading';
import { TaskCodeBadge } from '../task-code-badge';
import { Dialog, DialogClose, DialogContent, DialogHeader, DialogTitle } from '../ui/dialog';
import { ScrollArea } from '../ui/scroll-area';
import { TaskComments } from '../task-comments';
import { TaskActivityTimeline } from './task-activity-timeline';
import { TaskDetailsHeaderActions } from './task-details-header-actions';
import { TaskDetailsStateDialog } from './task-details-state-dialog';
import { TaskOverviewSection } from './task-overview-section';
import { useTaskDetails } from '@/hooks/use-task-details';

interface TaskDetailsProps {
  taskId: string;
  open: boolean;
  targetCommentId?: string;
  targetCommentCreatedAt?: string;
  onOpenChange: (open: boolean) => void;
  onEdit: () => void;
}

export const TaskDetails = ({
  taskId,
  open,
  targetCommentId,
  targetCommentCreatedAt,
  onOpenChange,
  onEdit,
}: TaskDetailsProps) => {
  const { task, isLoading, archive, isArchiving } = useTaskDetails(taskId, open, () => onOpenChange(false));

  if (isLoading) {
    return (
      <TaskDetailsStateDialog open={open} onOpenChange={onOpenChange}>
        <LoadingSpinner size="3rem" />
      </TaskDetailsStateDialog>
    );
  }

  if (!task) {
    return (
      <TaskDetailsStateDialog open={open} onOpenChange={onOpenChange}>
        <p className="text-sm text-slate-500 dark:text-slate-400">Task not found</p>
      </TaskDetailsStateDialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="flex max-h-[calc(100vh-5rem)] flex-col gap-0 overflow-y-auto p-0 sm:max-h-[calc(100vh-5rem)] sm:max-w-[calc(100vw-3rem)] xl:max-w-7xl xl:overflow-hidden 2xl:max-w-[96rem]"
        autoFocus={false}
        showCloseButton={false}
      >
        <DialogHeader className="bg-background flex flex-row items-center justify-between gap-4 border-b border-slate-200 px-6 py-5 dark:border-slate-700">
          <div className="min-w-0 space-y-2">
            <TaskCodeBadge code={task.code} className="max-w-full" />
            <DialogTitle className="truncate text-2xl font-semibold text-slate-900 dark:text-slate-100">
              {task.title}
            </DialogTitle>
          </div>
          <div className="flex items-center gap-2">
            <TaskDetailsHeaderActions
              status={task.status}
              isArchiving={isArchiving}
              onEdit={onEdit}
              onArchive={archive}
            />
            <DialogClose asChild>
              <button
                type="button"
                className="w-fit rounded-md p-2 font-medium transition-colors hover:bg-slate-100 dark:hover:bg-slate-800"
              >
                <X className="h-4 w-4 text-slate-500 dark:text-slate-400" />
              </button>
            </DialogClose>
          </div>
        </DialogHeader>

        <div className="grid grid-cols-1 xl:min-h-0 xl:flex-1 xl:grid-cols-[minmax(0,1.2fr)_minmax(24rem,0.9fr)] xl:overflow-hidden">
          <ScrollArea className="max-h-[80vh] pb-2">
            <div className="flex flex-col gap-6 px-6 py-6">
              <TaskOverviewSection task={task} />
              <TaskActivityTimeline updates={task.updates} />
            </div>
          </ScrollArea>

          <TaskComments
            taskId={taskId}
            projectId={task.project_id}
            open={open}
            targetCommentId={targetCommentId}
            targetCommentCreatedAt={targetCommentCreatedAt}
          />
        </div>
      </DialogContent>
    </Dialog>
  );
};

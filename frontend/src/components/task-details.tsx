import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Activity, Calendar, CircleCheck, ClockArrowUp, Pencil, Tag, Trash2, User, X } from 'lucide-react';
import { useState } from 'react';
import { LoadingSpinner } from './loading';
import { TaskDetailsUpdate } from './task-details-update';
import { TaskPriorityBadge } from './task-priority-badge';
import { TaskStatusBadge } from './task-status-badge';
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle } from './ui/dialog';
import { ScrollArea } from './ui/scroll-area';
import { archiveTask, getTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { cn } from '@/lib/utils';
import { sanitizeHTML } from '@/utils/html';

interface TaskDetailsProps {
  taskId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onEdit: () => void;
}

export const TaskDetails = ({ taskId, open, onOpenChange, onEdit }: TaskDetailsProps) => {
  const queryClient = useQueryClient();
  const [confirmingArchive, setConfirmingArchive] = useState(false);

  const { data: task, isLoading } = useQuery({
    queryKey: taskQueryKeys.details(taskId),
    queryFn: () => getTask(taskId),
    enabled: open,
  });

  const { mutate: archive, isPending: isArchiving } = useMutation({
    mutationFn: () => archiveTask(taskId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allGrouped() });
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allCounts() });
      onOpenChange(false);
    },
    onSettled: () => {
      setConfirmingArchive(false);
    },
  });

  const updates = task?.updates || [];

  if (isLoading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <div className="flex min-h-50 flex-col items-center justify-center">
            <LoadingSpinner size="3rem" />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  if (!task) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <div className="flex min-h-50 flex-col items-center justify-center">
            <p className="text-sm text-slate-500 dark:text-slate-400">Task not found</p>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  const infoTitleClassNames = 'text-sm font-medium text-slate-700 dark:text-slate-300';
  const infoValueClassNames = 'text-sm text-slate-500 dark:text-slate-400';

  return (
    <Dialog
      open={open}
      onOpenChange={(value) => {
        if (!value) setConfirmingArchive(false);
        onOpenChange(value);
      }}
    >
      <DialogContent className="pr-0 md:max-w-2xl" autoFocus={false} showCloseButton={false}>
        <DialogHeader className="position-sticky bg-background top-0 mr-4 flex flex-row items-center justify-between gap-2 border-b border-slate-200 pb-4 dark:border-slate-700">
          <DialogTitle className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{task.title}</DialogTitle>
          <div className="flex items-center gap-2">
            {confirmingArchive ? (
              <>
                <span className="text-xs text-slate-500 dark:text-slate-400">Archive task?</span>
                <button
                  type="button"
                  className="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950"
                  onClick={() => archive()}
                  disabled={isArchiving}
                >
                  {isArchiving ? 'Archiving...' : 'Confirm'}
                </button>
                <button
                  type="button"
                  className="rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
                  onClick={() => setConfirmingArchive(false)}
                >
                  Cancel
                </button>
              </>
            ) : (
              <>
                <button
                  type="button"
                  className="w-fit rounded-md p-2 font-medium transition-colors hover:bg-slate-100 dark:hover:bg-slate-800"
                  onClick={onEdit}
                >
                  <Pencil className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                </button>
                {task.status !== 'archived' && (
                  <button
                    type="button"
                    className="w-fit rounded-md p-2 font-medium transition-colors hover:bg-red-50 dark:hover:bg-red-950"
                    onClick={() => setConfirmingArchive(true)}
                  >
                    <Trash2 className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                  </button>
                )}
              </>
            )}
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

        <ScrollArea className="max-h-[50vh] pr-4">
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1">
              <h3 className="text-xs font-semibold tracking-wide text-slate-500 uppercase dark:text-slate-400">
                Description
              </h3>
              <DialogDescription asChild>
                <div dangerouslySetInnerHTML={{ __html: sanitizeHTML(task.description) }} />
              </DialogDescription>
            </div>
            <div className="grid grid-cols-2 gap-6 border-b border-slate-200 pb-4 dark:border-slate-700">
              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <CircleCheck className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Status</p>
                <TaskStatusBadge status={task.status} />
              </div>

              <div />

              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <User className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Author</p>
                <p className={cn(infoValueClassNames, 'col-start-2')}>{task.author.name}</p>
              </div>
              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <User className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Responsible</p>
                <p className={cn(infoValueClassNames, 'col-start-2')}>{task.responsible?.name || '-'}</p>
              </div>

              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <ClockArrowUp className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Priority</p>
                <TaskPriorityBadge priority={task.priority} />
              </div>

              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <Calendar className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Due date</p>
                <p className={cn(infoValueClassNames, 'col-start-2')}>{formatDateString(task.due_date)}</p>
              </div>
              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <Calendar className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Created at</p>
                <p className={cn(infoValueClassNames, 'col-start-2')}>{formatDateString(task.created_at)}</p>
              </div>
              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <Calendar className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Last updated</p>
                <p className={cn(infoValueClassNames, 'col-start-2')}>{formatDateString(task.updated_at)}</p>
              </div>

              <div className="col-span-2 grid grid-cols-[1rem_1fr] items-center gap-2">
                <Tag className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Tags</p>

                <div className="col-span-2 flex w-full flex-wrap gap-2">
                  {task.tags?.map((tag) => (
                    <div
                      key={tag}
                      className="w-fit rounded-sm border border-slate-200 px-2 py-0.5 text-xs font-medium text-slate-500 dark:border-slate-700 dark:text-slate-400"
                    >
                      {tag}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {
            <div className="flex w-full flex-col gap-4 overflow-y-auto pt-4 pr-2">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className="text-sm font-medium text-slate-700 dark:text-slate-300">
                  Activity timeline{' '}
                  <span className="text-xs text-slate-500 dark:text-slate-400">({updates.length})</span>
                </p>
              </div>
              <div className="flex flex-col gap-3">
                {updates.map((update) => (
                  <TaskDetailsUpdate key={update.id} update={update} />
                ))}
              </div>
            </div>
          }
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
};

const formatDateString = (date: string | null) => {
  if (!date) {
    return '-';
  }

  return new Date(date).toLocaleDateString([], {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

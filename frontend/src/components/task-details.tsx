import { useQuery } from '@tanstack/react-query';
import { Activity, Calendar, User } from 'lucide-react';
import { useMemo } from 'react';
import { Avatar } from './avatar';
import { LoadingSpinner } from './loading';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from './ui/dialog';
import { ScrollArea } from './ui/scroll-area';
import type { TaskStatus } from '@/types/task';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';
import { getTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { cn } from '@/lib/utils';

interface TaskDetailsProps {
  taskId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const TaskDetails = ({ taskId, open, onOpenChange }: TaskDetailsProps) => {
  const { data: task } = useQuery({
    queryKey: taskQueryKeys.details(taskId),
    queryFn: () => getTask(taskId),
    enabled: open,
  });

  const reversedChanges = useMemo(() => {
    if (!task?.changes) {
      return [];
    }

    return [...task.changes].reverse();
  }, [task]);

  if (!task) {
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

  const infoTitleClassNames = 'text-sm font-medium text-slate-700 dark:text-slate-300';
  const infoValueClassNames = 'text-sm text-slate-500 dark:text-slate-400';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="pr-0 md:max-w-2xl" autoFocus={false}>
        <DialogHeader className="position-sticky bg-background top-0 mr-4 border-b border-slate-200 pb-4 dark:border-slate-700">
          <DialogTitle className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{task.title}</DialogTitle>
          <TaskStatusLabel status={task.status} />
        </DialogHeader>

        <ScrollArea className="max-h-[50vh] pr-4">
          <div className="flex flex-col gap-4">
            <DialogDescription>{task.description}</DialogDescription>
            <div className="grid grid-cols-2 gap-4 border-b border-slate-200 pb-4 dark:border-slate-700">
              <div className="grid grid-cols-[1rem_1fr] items-center gap-2">
                <User className="h-4 w-4 text-slate-500 dark:text-slate-400" />
                <p className={infoTitleClassNames}>Author</p>
                <p className={cn(infoValueClassNames, 'col-start-2')}>{task.author.name}</p>
              </div>

              <div />

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
            </div>
          </div>

          <div className="flex w-full flex-col gap-4 overflow-y-auto pt-4 pr-2">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-slate-500 dark:text-slate-400" />
              <p className="text-sm font-medium text-slate-700 dark:text-slate-300">
                Activity timeline{' '}
                <span className="text-xs text-slate-500 dark:text-slate-400">({reversedChanges.length})</span>
              </p>
            </div>
            <div className="flex flex-col gap-3">
              {reversedChanges.map((change) => (
                <div key={change.id} className="flex w-full items-center gap-3">
                  {change.author?.name && <Avatar name={change.author.name} size="sm" />}

                  <div className="flex flex-1 items-start justify-between gap-8 rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-900">
                    <p className={infoValueClassNames}>{change.change_description}</p>
                    <p className={infoValueClassNames}>{formatRelativeActivityDateString(change.created_at)}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
};

const formatDateString = (date: string) => {
  return new Date(date).toLocaleDateString([], {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

const statusClassNames: Record<TaskStatus, string> = {
  pending: 'text-slate-900 dark:text-slate-500 bg-slate-500/10 border-slate-800',
  doing: 'text-blue-900 dark:text-blue-800 bg-blue-800/10 border-blue-800',
  done: 'text-emerald-900 dark:text-emerald-800 bg-emerald-800/10 border-emerald-800',
  archived: 'text-red-900 dark:text-red-800 bg-red-800/10 border-red-800',
};

const statusLabelText: Record<TaskStatus, string> = {
  pending: 'Pending',
  doing: 'Doing',
  done: 'Done',
  archived: 'Archived',
};

const TaskStatusLabel = ({ status }: { status: TaskStatus }) => {
  return (
    <div className={cn('w-fit rounded-full border px-2 py-0.5 text-xs font-medium', statusClassNames[status])}>
      {statusLabelText[status]}
    </div>
  );
};

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Activity, Calendar, ChevronDown, CircleCheck, ClockArrowUp, Pencil, Tag, Trash2, User, X } from 'lucide-react';
import { useState } from 'react';
import { LoadingSpinner } from './loading';
import { TaskComments } from './task-comments';
import { TaskDetailsUpdate } from './task-details-update';
import { TaskPriorityBadge } from './task-priority-badge';
import { TaskStatusBadge } from './task-status-badge';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from './ui/collapsible';
import { Dialog, DialogClose, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { ScrollArea } from './ui/scroll-area';
import { isHtmlContentEmpty, sanitizeHTML } from '@/utils/html';
import { formatDateString, formatTaskDueDate } from '@/utils/date';
import { archiveTask, getTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { cn } from '@/lib/utils';

interface TaskDetailsProps {
  taskId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onEdit: () => void;
}

const infoFieldClassNames = 'grid grid-cols-[1rem_1fr] items-center gap-2';
const infoTitleClassNames = 'text-sm font-medium text-slate-700 dark:text-slate-300';
const infoValueClassNames = 'text-sm text-slate-500 dark:text-slate-400 col-start-2';
const infoIconClassNames = 'h-4 w-4 text-slate-500 dark:text-slate-400';

export const TaskDetails = ({ taskId, open, onOpenChange, onEdit }: TaskDetailsProps) => {
  const queryClient = useQueryClient();
  const [confirmingArchive, setConfirmingArchive] = useState(false);
  const [activityOpen, setActivityOpen] = useState(false);

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

  const handleOpenChange = (value: boolean) => {
    if (!value) {
      setConfirmingArchive(false);
      setActivityOpen(false);
    }
    onOpenChange(value);
  };

  const updates = task?.updates || [];

  if (isLoading) {
    return (
      <Dialog open={open} onOpenChange={handleOpenChange}>
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
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent>
          <div className="flex min-h-50 flex-col items-center justify-center">
            <p className="text-sm text-slate-500 dark:text-slate-400">Task not found</p>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  const renderHeaderButtons = () => {
    if (confirmingArchive) {
      return (
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
      );
    }

    return (
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
    );
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        className="flex max-h-[calc(100vh-5rem)] flex-col gap-0 overflow-y-auto p-0 sm:max-h-[calc(100vh-5rem)] sm:max-w-6xl xl:overflow-hidden"
        autoFocus={false}
        showCloseButton={false}
      >
        <DialogHeader className="bg-background flex flex-row items-center justify-between gap-4 border-b border-slate-200 px-6 py-5 dark:border-slate-700">
          <div className="min-w-0">
            <DialogTitle className="truncate text-2xl font-semibold text-slate-900 dark:text-slate-100">
              {task.title}
            </DialogTitle>
          </div>
          <div className="flex items-center gap-2">
            {renderHeaderButtons()}
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
              <section className="flex flex-col gap-4">
                <div className="flex items-center justify-between gap-3">
                  <h3 className="text-xs font-semibold tracking-[0.16em] text-slate-500 uppercase dark:text-slate-400">
                    Overview
                  </h3>
                  <TaskStatusBadge status={task.status} />
                </div>

                <div className="rounded-2xl border border-slate-200 bg-slate-50/60 p-5 dark:border-slate-700 dark:bg-slate-900/40">
                  <div className="prose prose-slate dark:prose-invert max-w-none text-sm">
                    {!isHtmlContentEmpty(task.description) ? (
                      <div dangerouslySetInnerHTML={{ __html: sanitizeHTML(task.description) }} />
                    ) : (
                      <p className="m-0 text-slate-500 dark:text-slate-400">No description provided yet.</p>
                    )}
                  </div>
                </div>

                <div className="grid gap-4 rounded-2xl border border-slate-200 p-5 md:grid-cols-2 dark:border-slate-700">
                  <div className={infoFieldClassNames}>
                    <CircleCheck className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Status</p>
                    <TaskStatusBadge status={task.status} />
                  </div>

                  <div className={infoFieldClassNames}>
                    <ClockArrowUp className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Priority</p>
                    <TaskPriorityBadge priority={task.priority} />
                  </div>

                  <div className={infoFieldClassNames}>
                    <User className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Author</p>
                    <p className={infoValueClassNames}>{task.author.name}</p>
                  </div>

                  <div className={infoFieldClassNames}>
                    <User className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Responsible</p>
                    <p className={infoValueClassNames}>{task.responsible?.name || '-'}</p>
                  </div>

                  <div className={infoFieldClassNames}>
                    <Calendar className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Due date</p>
                    <p className={infoValueClassNames}>{formatTaskDueDate(task.due_date)}</p>
                  </div>

                  <div className={infoFieldClassNames}>
                    <Calendar className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Created at</p>
                    <p className={infoValueClassNames}>{formatDateString(task.created_at)}</p>
                  </div>

                  <div className={infoFieldClassNames}>
                    <Calendar className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Last updated</p>
                    <p className={infoValueClassNames}>{formatDateString(task.updated_at)}</p>
                  </div>

                  <div className={cn(infoFieldClassNames, 'md:col-span-2')}>
                    <Tag className={infoIconClassNames} />
                    <p className={infoTitleClassNames}>Tags</p>
                    <div className="col-span-2 flex w-full flex-wrap gap-2">
                      {task.tags?.length ? (
                        task.tags.map((tag) => (
                          <div
                            key={tag}
                            className="rounded-full border border-slate-200 px-2.5 py-1 text-xs font-medium text-slate-600 dark:border-slate-700 dark:text-slate-300"
                          >
                            {tag}
                          </div>
                        ))
                      ) : (
                        <p className={infoValueClassNames}>No tags added.</p>
                      )}
                    </div>
                  </div>
                </div>
              </section>

              <Collapsible
                open={activityOpen}
                onOpenChange={setActivityOpen}
                className="rounded-2xl border border-slate-200 dark:border-slate-700"
              >
                <CollapsibleTrigger className="flex w-full items-center justify-between gap-3 px-5 py-4 text-left transition-colors hover:bg-slate-50 dark:hover:bg-slate-900/50">
                  <div className="flex items-center gap-3">
                    <div className="rounded-full bg-slate-100 p-2 dark:bg-slate-800">
                      <Activity className="h-4 w-4 text-slate-600 dark:text-slate-300" />
                    </div>
                    <div>
                      <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">Activity timeline</p>
                      <p className="text-xs text-slate-500 dark:text-slate-400">
                        {updates.length ? `${updates.length} updates` : 'No activity yet'}
                      </p>
                    </div>
                  </div>
                  <ChevronDown
                    className={cn(
                      'h-4 w-4 text-slate-500 transition-transform dark:text-slate-400',
                      activityOpen && 'rotate-180',
                    )}
                  />
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <div className="border-t border-slate-200 px-5 py-4 dark:border-slate-700">
                    {updates.length > 0 ? (
                      <div className="flex flex-col">
                        {updates.map((update, index) => (
                          <TaskDetailsUpdate key={update.id} update={update} isLast={index === updates.length - 1} />
                        ))}
                      </div>
                    ) : (
                      <p className="text-sm text-slate-500 dark:text-slate-400">
                        Activity will appear here as the task changes.
                      </p>
                    )}
                  </div>
                </CollapsibleContent>
              </Collapsible>
            </div>
          </ScrollArea>

          <TaskComments taskId={taskId} projectId={task.project_id} open={open} />
        </div>
      </DialogContent>
    </Dialog>
  );
};

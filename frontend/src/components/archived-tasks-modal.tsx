import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Archive, ArchiveRestore, X } from 'lucide-react';
import { useMemo, useState } from 'react';
import { TaskDetails } from './task-details';
import { TaskPriorityBadge } from './task-priority-badge';
import { EditTask } from './task-form/edit-task';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { ScrollArea } from './ui/scroll-area';
import type { Task } from '@/types/task';
import type { Project } from '@/types/project';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { ProjectMemberRole } from '@/types/project';
import { listGroupedTasksByProjectId, updateTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { useAuth } from '@/hooks/use-auth';
import { DEFAULT_TASK_LIMIT } from '@/constants/tasks';

interface ArchivedTasksModalProps {
  project: Project;
}

export const ArchivedTasksModal = ({ project }: ArchivedTasksModalProps) => {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [editingTaskId, setEditingTaskId] = useState<string | null>(null);

  const queryClient = useQueryClient();
  const [pickingStatusForTaskId, setPickingStatusForTaskId] = useState<string | null>(null);

  const isCreator = project.members.some((m) => m.user_id === user?.id && m.role === ProjectMemberRole.Creator);

  const {
    mutate: restore,
    isPending: isRestoring,
    variables: restoreVariables,
  } = useMutation({
    mutationFn: ({ task, projectColumnId }: { task: Task; projectColumnId: string }) =>
      updateTask({
        id: task.id,
        title: task.title,
        description: task.description,
        project_column_id: projectColumnId,
        priority: task.priority,
        due_date: task.due_date,
        responsible_id: task.responsible_id,
        tags: task.tags ?? [],
      }),
    onSuccess: () => {
      setPickingStatusForTaskId(null);
      queryClient.invalidateQueries({
        queryKey: taskQueryKeys.listGroupedByProjectId({
          projectId: project.id,
          projectColumnIds: [],
          archived: true,
          limit: DEFAULT_TASK_LIMIT,
          taskOrder: '',
          updatedAt: null,
        }),
      });
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allGrouped() });
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allCounts() });
    },
  });

  const { data, fetchNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: taskQueryKeys.listGroupedByProjectId({
      projectId: project.id,
      projectColumnIds: [],
      archived: true,
      limit: DEFAULT_TASK_LIMIT,
      taskOrder: '',
      updatedAt: null,
    }),
    queryFn: ({ pageParam }) =>
      listGroupedTasksByProjectId({
        projectId: project.id,
        projectColumnIds: [],
        archived: true,
        taskOrder: pageParam.taskOrder,
        updatedAt: pageParam.updatedAt,
        limit: DEFAULT_TASK_LIMIT,
      }),
    getNextPageParam: (lastPage) => {
      const page = Object.values(lastPage)[0];
      if (!page.has_next) return undefined;
      const lastTask = page.data[page.data.length - 1];
      return { taskOrder: lastTask.order, updatedAt: lastTask.updated_at };
    },
    initialPageParam: { taskOrder: '', updatedAt: null as string | null },
    enabled: open,
  });

  const archivedTasks = useMemo(
    () => data?.pages.flatMap((page) => Object.values(page).flatMap((column) => column.data)) ?? [],
    [data],
  );

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
  });

  if (!isCreator) return null;

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:text-slate-500 dark:hover:bg-slate-800 dark:hover:text-slate-300"
      >
        <Archive className="h-3.5 w-3.5" />
        Archived
      </button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="md:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-base">
              <Archive className="h-4 w-4 text-slate-400" />
              Archived tasks
            </DialogTitle>
          </DialogHeader>

          <ScrollArea className="max-h-[60vh]">
            {archivedTasks.length === 0 && !isFetchingNextPage ? (
              <p className="py-8 text-center text-sm text-slate-400 dark:text-slate-500">No archived tasks</p>
            ) : (
              <div className="flex flex-col gap-1 pr-3">
                {archivedTasks.map((task) => {
                  const isPicking = pickingStatusForTaskId === task.id;
                  const isRestoringThis = isRestoring && restoreVariables.task.id === task.id;

                  return (
                    <div
                      key={task.id}
                      className="group flex w-full items-center gap-2 rounded-md px-3 py-2 transition-colors hover:bg-slate-50 dark:hover:bg-slate-800"
                    >
                      <button
                        type="button"
                        onClick={() => !isPicking && setSelectedTaskId(task.id)}
                        className="flex min-w-0 flex-1 items-center gap-3 text-left"
                      >
                        <span className="truncate text-sm text-slate-700 dark:text-slate-300">{task.title}</span>
                        {!isPicking && <TaskPriorityBadge priority={task.priority} />}
                      </button>

                      {isPicking ? (
                        <div className="flex shrink-0 items-center gap-1">
                          {project.columns.map((column) => (
                            <button
                              key={column.id}
                              type="button"
                              disabled={isRestoringThis}
                              onClick={() => restore({ task, projectColumnId: column.id! })}
                              className="rounded border px-2 py-0.5 text-xs font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                            >
                              {isRestoringThis && restoreVariables.projectColumnId === column.id ? '...' : column.name}
                            </button>
                          ))}
                          <button
                            type="button"
                            onClick={() => setPickingStatusForTaskId(null)}
                            className="ml-0.5 rounded p-0.5 text-slate-400 hover:bg-slate-200 hover:text-slate-600 dark:hover:bg-slate-700 dark:hover:text-slate-300"
                          >
                            <X className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      ) : (
                        <button
                          type="button"
                          title="Restore task"
                          onClick={() => setPickingStatusForTaskId(task.id)}
                          className="shrink-0 rounded p-1 text-slate-400 opacity-0 transition-opacity group-hover:opacity-100 hover:bg-slate-200 hover:text-slate-600 dark:hover:bg-slate-700 dark:hover:text-slate-300"
                        >
                          <ArchiveRestore className="h-3.5 w-3.5" />
                        </button>
                      )}
                    </div>
                  );
                })}
                <div ref={sentinelRef} className="h-1" aria-hidden="true" />
                {isFetchingNextPage && <p className="py-2 text-center text-xs text-slate-400">Loading...</p>}
              </div>
            )}
          </ScrollArea>
        </DialogContent>
      </Dialog>

      {selectedTaskId && (
        <TaskDetails
          taskId={selectedTaskId}
          open={!!selectedTaskId}
          onOpenChange={(value) => {
            if (!value) setSelectedTaskId(null);
          }}
          onEdit={() => {
            setEditingTaskId(selectedTaskId);
            setSelectedTaskId(null);
          }}
        />
      )}
      {editingTaskId && (
        <EditTask
          taskId={editingTaskId}
          open={!!editingTaskId}
          onOpenChange={(value) => {
            if (!value) setEditingTaskId(null);
          }}
        />
      )}
    </>
  );
};

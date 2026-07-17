import { ArchiveRestore, X } from 'lucide-react';
import { TaskPriorityBadge } from '../task-priority-badge';
import { useArchivedTasksList } from './use-archived-tasks-list';
import { useProjectDetails } from '@/hooks/use-project-details';

interface ArchivedTasksListProps {
  open: boolean;
  onSelectTask: (taskId: string) => void;
  projectId: string;
}

export const ArchivedTasksList = ({ projectId, open, onSelectTask }: ArchivedTasksListProps) => {
  const { data: project } = useProjectDetails(projectId, {
    enabled: open,
  });

  const columns = project?.columns.filter((column) => !!column.id) || [];
  const {
    archivedTasks,
    isFetchingNextPage,
    sentinelRef,
    pickingStatusForTaskId,
    restoringTaskId,
    startPickingStatus,
    restoreTaskToColumn,
    restoringProjectColumnId,
    stopPickingStatus,
  } = useArchivedTasksList(project?.id || '', open && !!project?.id);

  if (archivedTasks.length === 0 && !isFetchingNextPage) {
    return <p className="py-8 text-center text-sm text-slate-400 dark:text-slate-500">No archived tasks</p>;
  }

  return (
    <div className="flex flex-col gap-1 pr-3">
      {archivedTasks.map((task) => {
        const isPickingStatus = pickingStatusForTaskId === task.id;
        const isRestoringThisTask = restoringTaskId === task.id;

        return (
          <div
            key={task.id}
            className="group flex w-full items-center gap-2 rounded-md px-3 py-2 transition-colors hover:bg-slate-50 dark:hover:bg-slate-800"
          >
            <button
              type="button"
              onClick={() => {
                if (isPickingStatus) {
                  return;
                }

                onSelectTask(task.id);
              }}
              className="flex min-w-0 flex-1 items-center gap-3 text-left"
            >
              <span className="truncate text-sm text-slate-700 dark:text-slate-300">{task.title}</span>
              {!isPickingStatus && <TaskPriorityBadge priority={task.priority} />}
            </button>

            {!isPickingStatus && (
              <button
                type="button"
                title="Restore task"
                onClick={() => startPickingStatus(task.id)}
                className="shrink-0 rounded p-1 text-slate-400 opacity-0 transition-opacity group-hover:opacity-100 hover:bg-slate-200 hover:text-slate-600 dark:hover:bg-slate-700 dark:hover:text-slate-300"
              >
                <ArchiveRestore className="h-3.5 w-3.5" />
              </button>
            )}

            {isPickingStatus && (
              <div className="flex shrink-0 items-center gap-1">
                {columns.map((column) => (
                  <button
                    key={column.id}
                    type="button"
                    disabled={isRestoringThisTask}
                    onClick={() => restoreTaskToColumn(task, column.id)}
                    className="rounded border px-2 py-0.5 text-xs font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {isRestoringThisTask && restoringProjectColumnId === column.id ? '...' : column.name}
                  </button>
                ))}
                <button
                  type="button"
                  onClick={stopPickingStatus}
                  className="ml-0.5 rounded p-0.5 text-slate-400 hover:bg-slate-200 hover:text-slate-600 dark:hover:bg-slate-700 dark:hover:text-slate-300"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
            )}
          </div>
        );
      })}
      <div ref={sentinelRef} className="h-1" aria-hidden="true" />
      {isFetchingNextPage && <p className="py-2 text-center text-xs text-slate-400">Loading...</p>}
    </div>
  );
};

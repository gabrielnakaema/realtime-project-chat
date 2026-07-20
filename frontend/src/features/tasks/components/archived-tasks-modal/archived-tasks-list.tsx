import { ArchiveRestore, X } from 'lucide-react';
import { TaskPriorityBadge } from '../task-priority-badge';
import { useArchivedTasksList } from './use-archived-tasks-list';
import { useProjectDetails } from '@/features/projects/hooks/use-project-details';

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
    return <p className="text-muted-foreground py-8 text-center text-sm">No archived tasks</p>;
  }

  return (
    <div className="flex flex-col gap-1 pr-3">
      {archivedTasks.map((task) => {
        const isPickingStatus = pickingStatusForTaskId === task.id;
        const isRestoringThisTask = restoringTaskId === task.id;

        return (
          <div
            key={task.id}
            className="group hover:bg-muted flex w-full items-center gap-2 rounded-md px-3 py-2 transition-colors"
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
              <span className="text-foreground truncate text-sm">{task.title}</span>
              {!isPickingStatus && <TaskPriorityBadge priority={task.priority} />}
            </button>

            {!isPickingStatus && (
              <button
                type="button"
                title="Restore task"
                onClick={() => startPickingStatus(task.id)}
                className="text-muted-foreground hover:bg-accent hover:text-muted-foreground shrink-0 rounded p-1 opacity-0 transition-opacity group-hover:opacity-100"
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
                  className="text-muted-foreground hover:bg-accent hover:text-muted-foreground ml-0.5 rounded p-0.5"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
            )}
          </div>
        );
      })}
      <div ref={sentinelRef} className="h-1" aria-hidden="true" />
      {isFetchingNextPage && <p className="text-muted-foreground py-2 text-center text-xs">Loading...</p>}
    </div>
  );
};

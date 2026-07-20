import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { useCallback, useEffect, useRef } from 'react';
import { Plus } from 'lucide-react';
import { TaskCard } from './task-card';
import type {
  BaseEventPayload,
  DropTargetLocalizedData,
  ElementDragType,
} from '@atlaskit/pragmatic-drag-and-drop/dist/types/internal-types';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Column } from '@/features/tasks/types/board';
import { CreateTask } from '@/features/tasks/components/task-form/create-task';
import { ProjectColumnActions } from '@/features/projects/components/project-column-actions';
import { cn } from '@/lib/utils';
import { buildProjectColumnSurface } from '@/features/projects/utils/project-column-colors';
import { useMoveTask } from '@/features/tasks/hooks/use-move-task';
import { useBoardColumnTasks } from '@/features/tasks/hooks/use-board-column-tasks';
import { DEFAULT_TASK_LIMIT } from '@/features/tasks/constants/tasks';
import { sanitizeHTML } from '@/shared/utils/html';

interface BoardColumnProps {
  column: Column;
  canEditColumns: boolean;
}

export const BoardColumn = ({ column, canEditColumns }: BoardColumnProps) => {
  const columnRef = useRef<HTMLDivElement>(null);
  const scrollableRef = useRef<HTMLDivElement | null>(null);

  const mutation = useMoveTask();

  const {
    columnTasks: tasks,
    sentinelRef,
    queryResult,
  } = useBoardColumnTasks({
    projectId: column.project_id,
    columnId: column.columnId,
    limit: DEFAULT_TASK_LIMIT,
  });
  const { isLoading } = queryResult;

  const handleColumnDrop = useCallback(
    (args: BaseEventPayload<ElementDragType> & DropTargetLocalizedData) => {
      const { dropTargets = [] } = args.location.current;
      if (dropTargets.length === 1) {
        const { data: sourceData } = args.source;

        const sourceTaskId = sourceData.taskId as string;
        const tasksLength = tasks.length || 0;

        mutation.mutate({
          afterTaskId: tasksLength > 0 ? tasks[tasksLength - 1].id : null,
          projectId: column.project_id,
          projectColumnId: column.columnId,
          taskId: sourceTaskId,
        });
      }
    },
    [column, mutation, tasks],
  );

  useEffect(() => {
    const columnEl = columnRef.current;
    const scrollableEl = scrollableRef.current;
    if (!columnEl || !scrollableEl) {
      return () => {};
    }

    return combine(
      dropTargetForElements({
        element: columnEl,
        canDrop: () => true,
        onDrop: handleColumnDrop,
      }),
    );
  }, [handleColumnDrop]);

  const handleTaskDrop = useCallback(
    (edge: Edge | null, droppedTaskId: string, targetIndex: number) => {
      if (edge === 'bottom') {
        const task = tasks[targetIndex];
        mutation.mutate({
          afterTaskId: task.id,
          projectId: task.project_id,
          projectColumnId: task.project_column_id,
          taskId: droppedTaskId,
        });
      }

      if (edge === 'top') {
        if (targetIndex > 0) {
          const task = tasks[targetIndex - 1];
          mutation.mutate({
            afterTaskId: task.id,
            projectId: task.project_id,
            projectColumnId: task.project_column_id,
            taskId: droppedTaskId,
          });
        } else {
          mutation.mutate({
            afterTaskId: null,
            projectId: column.project_id,
            projectColumnId: column.columnId,
            taskId: droppedTaskId,
          });
        }
      }
    },
    [column, mutation, tasks],
  );

  const surface = buildProjectColumnSurface(column.color);

  return (
    <div key={column.id} className={cn('flex min-w-84 flex-1 shrink-0 flex-col overflow-hidden')} ref={columnRef}>
      <div className="mb-4 flex w-full flex-col gap-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="h-2 min-h-2 w-2 min-w-2 rounded-full" style={{ backgroundColor: surface.accentColor }} />
            <span
              className="inline-flex items-center rounded-lg px-2 py-0.5 text-[11px] font-semibold"
              style={{
                backgroundColor: surface.backgroundColor,
                color: surface.accentColor,
              }}
            >
              {column.total}
            </span>
            <h3 className="text-foreground text-sm font-semibold">{column.title}</h3>
          </div>
          <div className="flex items-center gap-3">
            {column.isDoneColumn && (
              <div className="bg-success/10 text-success w-fit rounded-sm px-1.5 py-0.5 text-[10px] font-bold tracking-[0.04em] uppercase">
                DONE
              </div>
            )}
            <ProjectColumnActions
              canEditColumn={canEditColumns}
              column={column}
              surfaceAccentColor={surface.accentColor}
            />
          </div>
        </div>

        <div className="flex w-full items-stretch gap-4">
          <div
            className="ml-1 h-full min-w-[2px]"
            style={{
              background: surface.borderColor,
            }}
          />

          <div
            className="prose line-clamp-2 truncate overflow-hidden text-xs break-all whitespace-pre-wrap italic"
            dangerouslySetInnerHTML={{
              __html: sanitizeHTML(column.description),
            }}
          />
        </div>
      </div>

      <div className="flex-1 space-y-3 overflow-y-auto" ref={scrollableRef}>
        {tasks.map((task, index) => (
          <TaskCard
            key={task.id}
            task={task}
            onDrop={(edge, droppedTaskId) => handleTaskDrop(edge, droppedTaskId, index)}
            columnSurface={surface}
          />
        ))}

        {tasks.length === 0 && !isLoading && (
          <p className="text-center text-xs font-medium tracking-tight">There are no tasks for this column.</p>
        )}

        <CreateTask
          projectId={column.project_id}
          initialProjectColumnId={column.id}
          trigger={
            <button
              type="button"
              className="mt-4 flex h-9 w-full cursor-pointer items-center justify-center gap-2 rounded-md border border-dashed text-xs font-semibold"
              style={{
                borderColor: surface.borderColor,
                color: surface.accentColor,
              }}
            >
              <Plus className="h-4 w-4" />
              Create task
            </button>
          }
        />
        <div ref={sentinelRef} className="h-1" aria-hidden="true" />
      </div>
    </div>
  );
};

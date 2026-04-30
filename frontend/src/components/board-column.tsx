import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { MoreHorizontal } from 'lucide-react';
import { useCallback, useEffect, useRef } from 'react';
import { TaskCard } from './task-card';
import type {
  BaseEventPayload,
  DropTargetLocalizedData,
  ElementDragType,
} from '@atlaskit/pragmatic-drag-and-drop/dist/types/internal-types';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Column } from '@/types/board';
import { cn } from '@/lib/utils';
import { buildProjectColumnSurface } from '@/lib/project-column-colors';
import { useMoveTask } from '@/hooks/use-move-task';
import { useBoardColumnTasks } from '@/hooks/use-board-column-tasks';
import { DEFAULT_TASK_LIMIT } from '@/constants/tasks';

interface BoardColumnProps {
  column: Column;
  onTaskClick?: (taskId: string) => void;
}

export const BoardColumn = ({ column, onTaskClick }: BoardColumnProps) => {
  const columnRef = useRef<HTMLDivElement>(null);
  const scrollableRef = useRef<HTMLDivElement | null>(null);

  const mutation = useMoveTask();

  const { columnTasks: tasks, sentinelRef } = useBoardColumnTasks({
    projectId: column.project_id,
    columnId: column.columnId,
    limit: DEFAULT_TASK_LIMIT,
  });

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
          projectColumnIds: [column.columnId],
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
          projectColumnIds: [task.project_column_id, column.columnId],
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
            projectColumnIds: [task.project_column_id, column.columnId],
            taskId: droppedTaskId,
          });
        } else {
          mutation.mutate({
            afterTaskId: null,
            projectId: column.project_id,
            projectColumnId: column.columnId,
            projectColumnIds: [column.columnId],
            taskId: droppedTaskId,
          });
        }
      }
    },
    [column, mutation, tasks],
  );

  const surface = buildProjectColumnSurface(column.color);

  return (
    <div
      key={column.id}
      className={cn('flex flex-col overflow-auto rounded-2xl border p-4 shadow-sm')}
      style={{ backgroundColor: surface.backgroundColor, borderColor: surface.borderColor }}
      ref={columnRef}
    >
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="font-semibold text-slate-900 dark:text-slate-100">{column.title}</h3>
          <span
            className="inline-flex items-center rounded-full border px-2 py-1 text-xs font-medium"
            style={{
              backgroundColor: surface.badgeBackground,
              borderColor: surface.borderColor,
              color: surface.accentColor,
            }}
          >
            {column.total}
          </span>
        </div>
        <button
          className="rounded p-1 text-slate-500 transition-colors hover:text-slate-700 dark:hover:text-slate-300"
          style={{ color: surface.accentColor }}
        >
          <MoreHorizontal className="h-4 w-4" />
        </button>
      </div>

      <div className="flex-1 space-y-3 overflow-y-auto" ref={scrollableRef}>
        {tasks.map((task, index) => (
          <TaskCard
            key={task.id}
            task={task}
            onDrop={(edge, droppedTaskId) => handleTaskDrop(edge, droppedTaskId, index)}
            onTaskClick={onTaskClick}
          />
        ))}
        <div ref={sentinelRef} className="h-1" aria-hidden="true" />
      </div>
    </div>
  );
};

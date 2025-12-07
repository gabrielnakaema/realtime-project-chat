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
import { useMoveTask } from '@/hooks/use-move-task';

interface BoardColumnProps {
  column: Column;
}

export const BoardColumn = ({ column }: BoardColumnProps) => {
  const columnRef = useRef<HTMLDivElement>(null);
  const scrollableRef = useRef<HTMLDivElement | null>(null);

  const mutation = useMoveTask();

  const handleColumnDrop = useCallback(
    (args: BaseEventPayload<ElementDragType> & DropTargetLocalizedData) => {
      const { dropTargets = [] } = args.location.current;
      if (dropTargets.length === 1) {
        const { data: sourceData } = args.source;

        const sourceTaskId = sourceData.taskId as string;
        const tasksLength = column.tasks.length || 0;

        mutation.mutate({
          afterTaskId: tasksLength > 0 ? column.tasks[tasksLength - 1].id : null,
          projectId: column.project_id,
          status: column.status,
          taskId: sourceTaskId,
        });
      }
    },
    [column, mutation],
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
        const task = column.tasks[targetIndex];
        mutation.mutate({
          afterTaskId: task.id,
          projectId: task.project_id,
          status: task.status,
          taskId: droppedTaskId,
        });
      }

      if (edge === 'top') {
        if (targetIndex > 0) {
          const task = column.tasks[targetIndex - 1];
          mutation.mutate({
            afterTaskId: task.id,
            projectId: task.project_id,
            status: task.status,
            taskId: droppedTaskId,
          });
        } else {
          mutation.mutate({
            afterTaskId: null,
            projectId: column.project_id,
            status: column.status,
            taskId: droppedTaskId,
          });
        }
      }
    },
    [column, mutation],
  );

  return (
    <div key={column.id} className={cn(column.color, 'flex flex-col overflow-auto rounded-lg p-4')} ref={columnRef}>
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="font-semibold text-slate-900 dark:text-slate-100">{column.title}</h3>
          <span className="inline-flex items-center rounded-full bg-slate-200 px-2 py-1 text-xs font-medium text-slate-700 dark:bg-slate-700 dark:text-slate-300">
            {column.total}
          </span>
        </div>
        <button className="rounded p-1 text-slate-500 hover:text-slate-700 dark:hover:text-slate-300">
          <MoreHorizontal className="h-4 w-4" />
        </button>
      </div>

      <div className="flex-1 space-y-3 overflow-y-auto" ref={scrollableRef}>
        {column.tasks.map((task, index) => (
          <TaskCard
            key={task.id}
            task={task}
            onDrop={(edge, droppedTaskId) => handleTaskDrop(edge, droppedTaskId, index)}
          />
        ))}
      </div>
    </div>
  );
};

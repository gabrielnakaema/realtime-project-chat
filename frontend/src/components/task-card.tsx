import { attachClosestEdge, extractClosestEdge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import { DropIndicator } from '@atlaskit/pragmatic-drag-and-drop-react-drop-indicator/box';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { draggable, dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { Calendar } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { Avatar } from './avatar';
import { TaskCodeBadge } from './task-code-badge';
import { TaskPriorityBadge } from './task-priority-badge';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Task } from '@/types/task';
import { useTaskDetailsRouting } from '@/hooks/use-task-details-routing';
import { cn } from '@/lib/utils';
import { sanitizeHTML } from '@/utils/html';

interface TaskCardProps {
  task: Task;
  onDrop: (edge: Edge | null, droppedTaskId: string) => void;
}

export const TaskCard = ({ task, onDrop }: TaskCardProps) => {
  const ref = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);
  const { openTask } = useTaskDetailsRouting();

  useEffect(() => {
    const el = ref.current;
    if (!el) {
      return () => {};
    }

    return combine(
      draggable({
        element: el,
        getInitialData: () => ({
          type: 'task',
          taskId: task.id,
          projectColumnId: task.project_column_id,
          title: task.title,
        }),
        onDragStart: () => setIsDragging(true),
        onDrop: () => setIsDragging(false),
      }),
      dropTargetForElements({
        element: el,
        canDrop: () => true,
        getData: ({ input, element }) => {
          const data = { type: 'task', taskId: task.id, projectColumnId: task.project_column_id, title: task.title };
          return attachClosestEdge(data, {
            input,
            element,
            allowedEdges: ['top', 'bottom'],
          });
        },
        getIsSticky: () => true,
        onDragEnter: ({ source, self }) => {
          if (source.data.taskId !== task.id) {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDrag: ({ source, self }) => {
          if (source.data.taskId !== task.id) {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDragLeave: () => {
          setClosestEdge(null);
        },
        onDrop: ({ source, self }) => {
          const draggedTaskId = source.data.taskId as string;
          const edge = extractClosestEdge(self.data);

          onDrop(edge, draggedTaskId);

          setClosestEdge(null);
        },
      }),
    );
  }, [task, onDrop]);

  return (
    <div className="relative px-0.5">
      <div
        ref={ref}
        className={cn(
          'cursor-pointer rounded-lg border border-slate-200 bg-white p-4 transition-shadow hover:shadow-md dark:border-slate-700 dark:bg-slate-900',
          isDragging && 'opacity-40 transition-opacity',
        )}
        onClick={() => openTask(task.id)}
      >
        <div className="pb-3">
          {task.code && <TaskCodeBadge code={task.code} className="mb-2" />}
          <div className="flex items-start justify-between gap-2">
            <h4 className="min-w-0 flex-1 text-sm leading-tight font-medium text-slate-900 dark:text-slate-100">
              {task.title}
            </h4>
            <TaskPriorityBadge priority={task.priority} />
          </div>
          {task.description && (
            <div
              className="mt-2 line-clamp-2 text-xs text-slate-600 dark:text-slate-400"
              dangerouslySetInnerHTML={{ __html: sanitizeHTML(task.description) }}
            />
          )}

          {!!task.tags?.length && (
            <div className="flex w-full flex-wrap items-center gap-2 pt-2">
              {task.tags.map((tag) => (
                <div
                  key={tag}
                  className="w-fit rounded-sm border border-slate-200 px-2 py-0.5 text-xs font-medium text-slate-500 dark:border-slate-700 dark:text-slate-400"
                >
                  {tag}
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">{task.responsible && <Avatar name={task.responsible.name} />}</div>

          {task.due_date && (
            <div className="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
              <Calendar className="h-3 w-3" />
              <span>{new Date(task.due_date).toLocaleDateString()}</span>
            </div>
          )}
        </div>
        {closestEdge && <DropIndicator edge={closestEdge} gap="12px" />}
      </div>
    </div>
  );
};

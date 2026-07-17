import { attachClosestEdge, extractClosestEdge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import { DropIndicator } from '@atlaskit/pragmatic-drag-and-drop-react-drop-indicator/box';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { draggable, dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { Calendar } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { TaskCodeBadge } from './task-code-badge';
import { TaskPriorityBadge } from './task-priority-badge';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Task } from '@/types/task';
import { cn } from '@/lib/utils';
import { useTaskDetailsRouting } from '@/hooks/use-task-details-routing';

interface TaskCardProps {
  task: Task;
  onDrop: (edge: Edge | null, droppedTaskId: string) => void;
  columnSurface: {
    backgroundColor: string;
    accentColor: string;
    badgeBackground: string;
    borderColor: string;
  };
}

const getResponsibleInitials = (responsible: string) => {
  const [first, last] = responsible.split(' ');
  return `${first.charAt(0)}${last ? last.charAt(0) : ''}`;
};

export const TaskCard = ({ task, onDrop, columnSurface }: TaskCardProps) => {
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
          'cursor-pointer rounded-lg border border-l-3 border-slate-200 bg-white p-3 transition-shadow hover:shadow-md dark:border-slate-800 dark:bg-slate-900',
          isDragging && 'opacity-40 transition-opacity',
        )}
        style={{
          borderLeftColor: columnSurface.accentColor,
        }}
        onClick={() => openTask(task.id)}
      >
        <div className="pb-3">
          <div className="flex flex-col items-start gap-2">
            <div className="flex items-center gap-2">
              <TaskPriorityBadge priority={task.priority} />
              {task.code && <TaskCodeBadge code={task.code} />}
            </div>
            <h4 className="min-w-0 flex-1 text-sm font-semibold text-slate-900 dark:text-slate-100">{task.title}</h4>
          </div>

          {!!task.tags?.length && (
            <div className="flex w-full flex-wrap items-center gap-2 pt-2">
              {task.tags.map((tag) => (
                <div
                  key={tag}
                  className="w-fit rounded-sm bg-slate-800 px-2 py-0.5 font-mono text-[10px] font-medium text-slate-600 dark:text-slate-400"
                >
                  {tag}
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1">
            <div className="flex min-h-5 min-w-5 items-center justify-center rounded-sm border p-1 font-mono text-[9px] leading-1 font-semibold uppercase dark:border-slate-700 dark:text-slate-300">
              {task.responsible ? getResponsibleInitials(task.responsible.name) : '-'}
            </div>
            <p className="font-sans text-[11px] leading-1.5 font-normal dark:text-slate-300">
              {task.responsible ? task.responsible.name : 'Unassigned'}
            </p>
          </div>

          {task.due_date && (
            <div className="flex items-center gap-1 font-sans text-[11px] font-medium text-slate-500 dark:text-slate-300">
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

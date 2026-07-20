import { attachClosestEdge, extractClosestEdge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import { DropIndicator } from '@atlaskit/pragmatic-drag-and-drop-react-drop-indicator/box';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { draggable, dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { CalendarDays, GripVertical } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { PreviewPriority, PreviewTask } from '@/features/landing/data/preview-board';
import { cn } from '@/lib/utils';

const priorityClasses: Record<PreviewPriority, string> = {
  Low: 'bg-success/10 text-success',
  Medium: 'bg-warning/10 text-warning',
  High: 'bg-destructive/10 text-destructive',
};

interface PreviewTaskCardProps {
  task: PreviewTask;
  accentColor: string;
  onDrop: (edge: Edge | null, droppedTaskId: string) => void;
}

export const PreviewTaskCard = ({ task, accentColor, onDrop }: PreviewTaskCardProps) => {
  const ref = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) {
      return () => {};
    }

    return combine(
      draggable({
        element: el,
        getInitialData: () => ({ type: 'preview-task', taskId: task.id }),
        onDragStart: () => setIsDragging(true),
        onDrop: () => setIsDragging(false),
      }),
      dropTargetForElements({
        element: el,
        canDrop: ({ source }) => source.data.type === 'preview-task',
        getData: ({ input, element }) =>
          attachClosestEdge(
            { type: 'preview-task', taskId: task.id },
            { input, element, allowedEdges: ['top', 'bottom'] },
          ),
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
        onDragLeave: () => setClosestEdge(null),
        onDrop: ({ source, self }) => {
          onDrop(extractClosestEdge(self.data), source.data.taskId as string);
          setClosestEdge(null);
        },
      }),
    );
  }, [task.id, onDrop]);

  return (
    <div className="relative">
      <article
        ref={ref}
        className={cn(
          'border-border bg-card group cursor-grab rounded-lg border border-l-3 p-3 transition-shadow hover:shadow-md active:cursor-grabbing',
          task.highlighted && 'ring-primary/60 ring-2',
          isDragging && 'opacity-40',
        )}
        style={{ borderLeftColor: accentColor }}
      >
        <div className="mb-2 flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className={cn('rounded px-1.5 py-1 text-[9px] font-bold uppercase', priorityClasses[task.priority])}>
              {task.priority}
            </span>
            <span className="text-muted-foreground font-mono text-[10px] font-semibold">{task.code}</span>
          </div>
          <GripVertical className="text-muted-foreground/50 size-3.5 opacity-0 transition-opacity group-hover:opacity-100" />
        </div>
        <h4 className="text-foreground text-sm font-semibold">{task.title}</h4>
        <div className="mt-3 flex flex-wrap gap-1.5">
          {task.tags.map((tag) => (
            <span key={tag} className="bg-muted text-muted-foreground rounded px-1.5 py-1 font-mono text-[9px]">
              {tag}
            </span>
          ))}
        </div>
        <div className="text-muted-foreground mt-4 flex items-center justify-between text-[10px]">
          <span className="flex items-center gap-1.5">
            <span className="border-border flex size-5 items-center justify-center rounded border font-mono text-[8px] font-semibold">
              {task.owner}
            </span>
            {task.ownerName}
          </span>
          {task.dueDate && (
            <span className="flex items-center gap-1">
              <CalendarDays className="size-3" /> {task.dueDate}
            </span>
          )}
        </div>
      </article>
      {closestEdge && <DropIndicator edge={closestEdge} gap="12px" />}
    </div>
  );
};

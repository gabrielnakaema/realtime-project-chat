import { dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { Check, Circle } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { PreviewTaskCard } from './preview-task-card';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { PreviewColumn, PreviewTask } from '@/features/landing/data/preview-board';
import { cn } from '@/lib/utils';

interface PreviewBoardColumnProps {
  column: PreviewColumn;
  tasks: ReadonlyArray<PreviewTask>;
  onTaskDrop: (edge: Edge | null, droppedTaskId: string, targetTaskId: string | null) => void;
}

export const PreviewBoardColumn = ({ column, tasks, onTaskDrop }: PreviewBoardColumnProps) => {
  const ref = useRef<HTMLDivElement>(null);
  const [isDraggedOver, setIsDraggedOver] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) {
      return () => {};
    }

    return dropTargetForElements({
      element: el,
      canDrop: ({ source }) => source.data.type === 'preview-task',
      onDrop: ({ source, location }) => {
        if (location.current.dropTargets[0]?.element === el) {
          onTaskDrop(null, source.data.taskId as string, null);
        }
        setIsDraggedOver(false);
      },
      onDragEnter: () => setIsDraggedOver(true),
      onDragLeave: () => setIsDraggedOver(false),
    });
  }, [onTaskDrop]);

  return (
    <div
      ref={ref}
      className={cn('min-h-[360px] rounded-xl p-3 transition-colors', isDraggedOver ? 'bg-muted/40' : 'bg-transparent')}
    >
      <div className="mb-4 flex items-center justify-between px-1">
        <div className="flex items-center gap-2">
          {column.isDone ? (
            <Check className="size-3.5" style={{ color: column.color }} />
          ) : (
            <Circle className="size-3 fill-current" style={{ color: column.color }} />
          )}
          <h3 className="text-foreground text-xs font-semibold">{column.name}</h3>
          <span className="bg-background/70 text-muted-foreground rounded px-1.5 py-0.5 font-mono text-[10px]">
            {tasks.length}
          </span>
        </div>
        <span className="text-muted-foreground text-xs" aria-hidden="true">
          ···
        </span>
      </div>

      <div className="space-y-3">
        {tasks.map((task) => (
          <PreviewTaskCard
            key={task.id}
            task={task}
            accentColor={column.color}
            onDrop={(edge, droppedTaskId) => onTaskDrop(edge, droppedTaskId, task.id)}
          />
        ))}
      </div>
    </div>
  );
};

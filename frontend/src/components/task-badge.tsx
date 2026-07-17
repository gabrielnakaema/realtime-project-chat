import type { TaskPriority, TaskStatus } from '@/types/task';
import { cn } from '@/lib/utils';

export type TaskBadgeColor = 'blue' | 'green' | 'yellow' | 'red' | 'slate';

interface TaskBadgeProps {
  children: React.ReactNode;
  color: TaskBadgeColor;
  className?: string;
  style?: React.CSSProperties;
}

const colorsClassNames: Record<string, string> = {
  blue: 'bg-blue-500/10  text-blue-500',
  green: 'bg-green-500/10 text-green-500',
  yellow: 'bg-yellow-500/10  text-yellow-500',
  red: 'bg-red-500/10  text-red-400',
  slate: 'bg-slate-500/10  text-slate-500',
};

export const priorityToColor: Record<TaskPriority, TaskBadgeColor> = {
  low: 'green',
  medium: 'yellow',
  high: 'red',
};

export const statusToColor: Record<TaskStatus, TaskBadgeColor> = {
  pending: 'slate',
  doing: 'blue',
  done: 'green',
  archived: 'red',
};

export const TaskBadge = ({ children, color, className, style }: TaskBadgeProps) => {
  return (
    <div
      className={cn(
        'w-fit rounded-sm px-1.5 py-0.5 text-[9px] font-bold tracking-[0.04em] uppercase',
        colorsClassNames[color],
        className,
      )}
      style={style}
    >
      {children}
    </div>
  );
};

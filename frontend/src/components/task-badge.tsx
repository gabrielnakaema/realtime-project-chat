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
  blue: 'bg-blue-500/10 border-blue-600 text-blue-600',
  green: 'bg-green-500/10 border-green-600 text-green-600',
  yellow: 'bg-yellow-500/10 border-yellow-600 text-yellow-600',
  red: 'bg-red-500/10 border-red-600 text-red-600',
  slate: 'bg-slate-500/10 border-slate-600 text-slate-600',
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
      className={cn('w-fit rounded-md border px-2 py-0.5 text-xs font-medium', colorsClassNames[color], className)}
      style={style}
    >
      {children}
    </div>
  );
};

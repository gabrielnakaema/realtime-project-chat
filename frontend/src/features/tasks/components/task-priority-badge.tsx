import { TaskBadge } from './task-badge';
import type { TaskBadgeColor } from './task-badge';
import type { TaskPriority } from '@/features/tasks/types/task';

export const TaskPriorityBadge = ({ priority }: { priority: TaskPriority }) => {
  return <TaskBadge color={priorityToColor[priority]}>{priorityLabelText[priority]}</TaskBadge>;
};

const priorityToColor: Record<TaskPriority, TaskBadgeColor> = {
  low: 'green',
  medium: 'yellow',
  high: 'red',
};

const priorityLabelText: Record<TaskPriority, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
};

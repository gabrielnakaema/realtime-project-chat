import { TaskBadge } from './task-badge';
import type { TaskBadgeColor } from './task-badge';
import type { TaskStatus } from '@/types/task';

const statusToColor: Record<TaskStatus, TaskBadgeColor> = {
  pending: 'slate',
  doing: 'blue',
  done: 'green',
  archived: 'red',
};

const statusLabelText: Record<TaskStatus, string> = {
  pending: 'Pending',
  doing: 'Doing',
  done: 'Done',
  archived: 'Archived',
};

export const TaskStatusBadge = ({ status }: { status: TaskStatus }) => {
  return <TaskBadge color={statusToColor[status]}>{statusLabelText[status]}</TaskBadge>;
};

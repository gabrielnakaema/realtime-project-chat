import { TaskBadge } from './task-badge';
import type { TaskBadgeColor } from './task-badge';
import type { TaskStatus } from '@/features/tasks/types/task';
import { buildProjectColumnSurface } from '@/features/projects/utils/project-column-colors';

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

export const TaskStatusBadge = ({
  status,
  label,
  color,
}: {
  status: TaskStatus;
  label?: string;
  color?: string | null;
}) => {
  if (color && label) {
    const surface = buildProjectColumnSurface(color);

    return (
      <TaskBadge
        color="slate"
        style={{
          backgroundColor: surface.badgeBackground,
          borderColor: surface.borderColor,
          color: surface.accentColor,
        }}
      >
        {label}
      </TaskBadge>
    );
  }

  return <TaskBadge color={statusToColor[status]}>{label || statusLabelText[status] || status}</TaskBadge>;
};

import { TaskBadge } from './task-badge';
import { buildProjectColumnSurface } from '@/features/projects/utils/project-column-colors';

export const TaskColumnBadge = ({ label, color }: { label?: string | null; color?: string | null }) => {
  if (!label) {
    return null;
  }

  if (!color) {
    return <TaskBadge color="slate">{label}</TaskBadge>;
  }

  const surface = buildProjectColumnSurface(color);

  return (
    <TaskBadge
      color="slate"
      style={{
        backgroundColor: surface.badgeBackground,
        borderColor: surface.borderColor,
        color: surface.accentColor,
      }}
      className="overflow-hidden whitespace-nowrap"
    >
      {label}
    </TaskBadge>
  );
};

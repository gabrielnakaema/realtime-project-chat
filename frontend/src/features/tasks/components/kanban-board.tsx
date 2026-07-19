import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { BoardColumn } from './board-column';
import type { Project } from '@/features/projects/types/project';
import { ProjectMemberRole } from '@/features/projects/types/project';
import { useRealtimeTaskSync } from '@/features/tasks/hooks/use-realtime-task-sync';
import { countTasksByColumn } from '@/features/tasks/services/tasks';
import { taskQueryKeys } from '@/shared/services/query-keys';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { getDefaultProjectColumnColor } from '@/features/projects/utils/project-column-colors';

interface KanbanBoardProps {
  project: Project;
}

export const KanbanBoard = ({ project }: KanbanBoardProps) => {
  const projectId = project.id;
  const projectColumnIds = useMemo(() => {
    return project.columns.map((column) => column.id).filter(Boolean);
  }, [project.columns]);

  useRealtimeTaskSync(projectId);

  const { data: countData } = useQuery({
    queryKey: taskQueryKeys.countsFor(projectId, projectColumnIds),
    queryFn: () => countTasksByColumn(projectId, projectColumnIds),
  });

  const { user } = useAuth();

  const columns = useMemo(() => {
    return project.columns.map((column, index) => ({
      id: column.id,
      color: column.color || getDefaultProjectColumnColor(index),
      title: column.name,
      description: column.description,
      columnId: column.id,
      isDoneColumn: column.is_done_column,
      project_id: project.id,
      total: countData?.[column.id] || 0,
    }));
  }, [project.columns, project.id, countData]);

  const isOwner = useMemo(() => {
    return project.members.some((member) => member.user_id === user?.id && member.role === ProjectMemberRole.Creator);
  }, [project.members, user?.id]);

  return (
    <div className="h-full">
      <div className="flex h-[calc(100vh-100px)] w-full max-w-full gap-6 overflow-x-auto pb-3">
        {columns.map((column) => (
          <BoardColumn canEditColumns={isOwner} column={column} key={column.id} />
        ))}
      </div>
    </div>
  );
};

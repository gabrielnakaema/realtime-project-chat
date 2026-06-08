import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ArchivedTasksModal } from './archived-tasks-modal';
import { BoardColumn } from './board-column';
import { CreateTask } from './task-form/create-task';
import type { Project } from '@/types/project';
import { ProjectMemberRole } from '@/types/project';
import { useRealtimeTaskSync } from '@/hooks/use-realtime-task-sync';
import { countTasksByColumn } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { useAuth } from '@/hooks/use-auth';
import { getDefaultProjectColumnColor } from '@/lib/project-column-colors';

interface KanbanBoardProps {
  project: Project;
}

export const KanbanBoard = ({ project }: KanbanBoardProps) => {
  const projectId = project.id;
  const projectColumnIds = useMemo(() => {
    return project.columns.map((column) => column.id).filter(Boolean);
  }, [project.columns]);

  useRealtimeTaskSync(projectId, projectColumnIds);

  const { data: countData } = useQuery({
    queryKey: taskQueryKeys.countByColumn(projectId, projectColumnIds),
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
      <div className="flex items-center justify-between pb-6">
        <div className="flex items-center gap-3">
          <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Task Board</h2>
          {isOwner && <ArchivedTasksModal project={project} />}
        </div>
        <CreateTask projectId={projectId} initialProjectColumnId={project.columns[0]?.id} />
      </div>

      <div className="flex h-[calc(100vh-200px)] w-full max-w-full gap-6 overflow-x-auto pb-3">
        {columns.map((column) => (
          <BoardColumn canEditColumns={isOwner} column={column} key={column.id} />
        ))}
      </div>
    </div>
  );
};

import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ArchivedTasksModal } from './archived-tasks-modal';
import { BoardColumn } from './board-column';
import { CreateTask } from './create-task';
import type { TaskStatus } from '@/types/task';
import type { Project } from '@/types/project';
import { ProjectMemberRole } from '@/types/project';
import { useRealtimeTaskSync } from '@/hooks/use-realtime-task-sync';
import { countTasksByStatus } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { TASK_STATUSES } from '@/constants/tasks';
import { useAuth } from '@/hooks/use-auth';

interface Column {
  id: string;
  title: string;
  status: TaskStatus;
  color: string;
}

const DEFAULT_COLUMNS: Column[] = [
  { id: 'pending', title: 'Pending', status: 'pending', color: 'bg-slate-100 dark:bg-slate-800' },
  { id: 'doing', title: 'Doing', status: 'doing', color: 'bg-blue-50 dark:bg-blue-950' },
  { id: 'done', title: 'Done', status: 'done', color: 'bg-emerald-50 dark:bg-emerald-950' },
];

interface KanbanBoardProps {
  project: Project;
  onTaskClick?: (taskId: string) => void;
}

export const KanbanBoard = ({ project, onTaskClick }: KanbanBoardProps) => {
  const projectId = project.id;

  useRealtimeTaskSync(projectId);

  const { data: countData } = useQuery({
    queryKey: taskQueryKeys.countByStatus(projectId, [...TASK_STATUSES]),
    queryFn: () => countTasksByStatus(projectId, [...TASK_STATUSES]),
  });

  const { user } = useAuth();

  const columns = useMemo(() => {
    return DEFAULT_COLUMNS.map((column) => ({
      id: column.id,
      color: column.color,
      title: column.title,
      status: column.status,
      project_id: project.id,
      total: countData?.[column.status] || 0,
    }));
  }, [project.id, countData]);

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
        <CreateTask projectId={projectId} projectMembers={project.members} />
      </div>

      <div className="grid h-[calc(100vh-200px)] grid-cols-1 gap-6 md:grid-cols-3">
        {columns.map((column) => (
          <BoardColumn column={column} key={column.id} onTaskClick={onTaskClick} />
        ))}
      </div>
    </div>
  );
};

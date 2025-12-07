import { useMemo } from 'react';
import { BoardColumn } from './board-column';
import { CreateTask } from './create-task';
import type { TaskStatus } from '@/types/task';
import type { Project } from '@/types/project';
import { useProjectTasks } from '@/hooks/use-project-tasks';

interface Column {
  id: string;
  title: string;
  status: TaskStatus;
  color: string;
}

const columns: Column[] = [
  { id: 'pending', title: 'Pending', status: 'pending', color: 'bg-slate-100 dark:bg-slate-800' },
  { id: 'doing', title: 'Doing', status: 'doing', color: 'bg-blue-50 dark:bg-blue-950' },
  { id: 'done', title: 'Done', status: 'done', color: 'bg-emerald-50 dark:bg-emerald-950' },
];

export const KanbanBoard = ({ project }: { project: Project }) => {
  const projectId = project.id;

  const { data, countData } = useProjectTasks(projectId);

  const columnTasks = useMemo(() => {
    return columns.map((column) => ({
      id: column.id,
      color: column.color,
      title: column.title,
      tasks: data?.[column.status]?.data || [],
      status: column.status,
      project_id: project.id,
      total: countData?.[column.status] || 0,
    }));
  }, [data, project.id, countData]);

  return (
    <div className="h-full">
      <div className="flex items-center justify-between pb-6">
        <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Task Board</h2>
        <CreateTask projectId={projectId} projectMembers={project.members} />
      </div>

      <div className="grid h-[calc(100vh-200px)] grid-cols-1 gap-6 md:grid-cols-3">
        {columnTasks.map((column) => (
          <BoardColumn column={column} key={column.id} />
        ))}
      </div>
    </div>
  );
};

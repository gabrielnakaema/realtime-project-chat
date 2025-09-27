import { TaskCard } from '@/components/task-card';
import { taskQueryKeys } from '@/services/query-keys';
import { updateTask } from '@/services/tasks';
import type { Task, TaskStatus } from '@/types/task';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { MoreHorizontal } from 'lucide-react';
import { useMemo, useState } from 'react';
import { CreateTask } from './create-task';
import { cn } from '@/lib/utils';
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

export function KanbanBoard({ projectId }: { projectId: string }) {
  const [draggedTask, setDraggedTask] = useState<Task | null>(null);
  const queryClient = useQueryClient();

  const { tasks } = useProjectTasks(projectId);

  const { mutate: mutateUpdateTask } = useMutation({
    mutationFn: updateTask,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.listByProjectId(projectId) });
    },
  });

  const handleDragStart = (task: Task) => {
    setDraggedTask(task);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = (e: React.DragEvent, newStatus: TaskStatus) => {
    e.preventDefault();

    if (draggedTask) {
      mutateUpdateTask({
        id: draggedTask.id,
        status: newStatus,
        title: draggedTask.title,
        description: draggedTask.description,
      });
    }

    setDraggedTask(null);
  };

  const tasksByStatus = useMemo(() => {
    return tasks.reduce(
      (acc, task) => {
        return {
          ...acc,
          [task.status]: [...acc[task.status], task],
        };
      },
      {
        pending: [],
        doing: [],
        done: [],
        archived: [],
      } as Record<TaskStatus, Task[]>,
    );
  }, [tasks]);

  return (
    <div className="h-full">
      <div className="flex items-center justify-between pb-6">
        <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Task Board</h2>
        <CreateTask projectId={projectId} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 h-[calc(100vh-200px)]">
        {columns.map((column) => {
          const columnTasks = tasksByStatus[column.status];
          return (
            <div
              key={column.id}
              className={cn(column.color, 'rounded-lg p-4 flex flex-col')}
              onDragOver={handleDragOver}
              onDrop={(e) => handleDrop(e, column.status)}
            >
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold text-slate-900 dark:text-slate-100">{column.title}</h3>
                  <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-full">
                    {columnTasks.length}
                  </span>
                </div>
                <button className="p-1 text-slate-500 hover:text-slate-700 dark:hover:text-slate-300 rounded">
                  <MoreHorizontal className="w-4 h-4" />
                </button>
              </div>

              <div className="flex-1 space-y-3 overflow-y-auto">
                {columnTasks.map((task) => (
                  <TaskCard key={task.id} task={task} onDragStart={() => handleDragStart(task)} />
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

import { Calendar } from 'lucide-react';
import { useState } from 'react';
import { Avatar } from './avatar';
import { TaskDetails } from './task-details';
import type { Task } from '@/types/task';

interface TaskCardProps {
  task: Task;
  onDragStart: () => void;
}

export function TaskCard({ task, onDragStart }: TaskCardProps) {
  const [open, setOpen] = useState(false);

  return (
    <>
      <div
        className="cursor-pointer rounded-lg border border-slate-200 bg-white p-4 transition-shadow hover:shadow-md dark:border-slate-700 dark:bg-slate-900"
        draggable
        onDragStart={onDragStart}
        onClick={() => setOpen(true)}
      >
        <div className="pb-3">
          <div className="flex items-start justify-between">
            <h4 className="text-sm leading-tight font-medium text-slate-900 dark:text-slate-100">{task.title}</h4>
          </div>
          {task.description && (
            <p className="mt-2 line-clamp-2 text-xs text-slate-600 dark:text-slate-400">{task.description}</p>
          )}
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">{task.author.name && <Avatar name={task.author.name} />}</div>

          {task.created_at && (
            <div className="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
              <Calendar className="h-3 w-3" />
              <span>{new Date(task.created_at).toLocaleDateString()}</span>
            </div>
          )}
        </div>
      </div>
      <TaskDetails taskId={task.id} open={open} onOpenChange={setOpen} />
    </>
  );
}

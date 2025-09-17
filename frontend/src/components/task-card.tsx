import type { Task } from '@/types/task';
import { Calendar } from 'lucide-react';

interface TaskCardProps {
  task: Task;
  onDragStart: () => void;
}

export function TaskCard({ task, onDragStart }: TaskCardProps) {
  return (
    <div
      className="cursor-move hover:shadow-md transition-shadow bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg p-4"
      draggable
      onDragStart={onDragStart}
    >
      <div className="pb-3">
        <div className="flex items-start justify-between">
          <h4 className="font-medium text-slate-900 dark:text-slate-100 text-sm leading-tight">{task.title}</h4>
        </div>
        {task.description && (
          <p className="text-xs text-slate-600 dark:text-slate-400 mt-2 line-clamp-2">{task.description}</p>
        )}
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {task.author && (
            <div className="w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center text-white text-xs font-medium">
              {task.author.name.charAt(0).toUpperCase()}
            </div>
          )}
        </div>

        {task.created_at && (
          <div className="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
            <Calendar className="w-3 h-3" />
            <span>{new Date(task.created_at).toLocaleDateString()}</span>
          </div>
        )}
      </div>
    </div>
  );
}

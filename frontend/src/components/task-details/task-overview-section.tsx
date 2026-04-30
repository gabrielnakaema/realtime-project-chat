import { TaskStatusBadge } from '../task-status-badge';
import { TaskDataGrid } from './task-data-grid';
import type { Task } from '@/types/task';
import { isHtmlContentEmpty, sanitizeHTML } from '@/utils/html';

interface TaskOverviewSectionProps {
  task: Task;
}

export const TaskOverviewSection = ({ task }: TaskOverviewSectionProps) => {
  return (
    <section className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <h3 className="text-xs font-semibold tracking-[0.16em] text-slate-500 uppercase dark:text-slate-400">
          Overview
        </h3>
        <TaskStatusBadge status={task.status} label={task.project_column?.name} color={task.project_column?.color} />
      </div>

      <div className="rounded-2xl border border-slate-200 bg-slate-50/60 p-5 dark:border-slate-700 dark:bg-slate-900/40">
        <div className="prose prose-slate dark:prose-invert max-w-none text-sm">
          {isHtmlContentEmpty(task.description) ? (
            <p className="m-0 text-slate-500 dark:text-slate-400">No description provided yet.</p>
          ) : (
            <div dangerouslySetInnerHTML={{ __html: sanitizeHTML(task.description) }} />
          )}
        </div>
      </div>

      <TaskDataGrid task={task} />
    </section>
  );
};

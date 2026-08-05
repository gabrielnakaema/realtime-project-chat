import { TaskColumnBadge } from '../task-column-badge';
import { TaskDataGrid } from './task-data-grid';
import type { Task } from '@/features/tasks/types/task';
import { isHtmlContentEmpty, richTextListClassName, sanitizeHTML } from '@/shared/utils/html';
import { cn } from '@/lib/utils';

interface TaskOverviewSectionProps {
  task: Task;
}

export const TaskOverviewSection = ({ task }: TaskOverviewSectionProps) => {
  return (
    <section className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-3">
        <h3 className="text-muted-foreground text-xs font-semibold tracking-[0.16em] uppercase">Overview</h3>
        <TaskColumnBadge label={task.project_column?.name} color={task.project_column?.color} />
      </div>

      <div className="border-border bg-muted/60 rounded-2xl border p-5">
        <div className="prose prose-slate dark:prose-invert max-w-none text-sm">
          {isHtmlContentEmpty(task.description) ? (
            <p className="text-muted-foreground m-0">No description provided yet.</p>
          ) : (
            <div
              className={cn(richTextListClassName)}
              dangerouslySetInnerHTML={{ __html: sanitizeHTML(task.description) }}
            />
          )}
        </div>
      </div>

      <TaskDataGrid task={task} />
    </section>
  );
};

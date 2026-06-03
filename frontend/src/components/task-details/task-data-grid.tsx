import { Calendar, CircleCheck, ClockArrowUp, Hash, Tag, User } from 'lucide-react';
import { TaskCodeBadge } from '../task-code-badge';
import { TaskPriorityBadge } from '../task-priority-badge';
import { TaskStatusBadge } from '../task-status-badge';
import type { Task } from '@/types/task';
import { formatDateString, formatTaskDueDate } from '@/utils/date';
import { cn } from '@/lib/utils';

interface TaskDataGridProps {
  task: Task;
}

const infoFieldClassNames = 'grid grid-cols-[1rem_1fr] items-center gap-2';
const infoTitleClassNames = 'text-sm font-medium text-slate-700 dark:text-slate-300';
const infoValueClassNames = 'text-sm text-slate-500 dark:text-slate-400 col-start-2';
const infoIconClassNames = 'h-4 w-4 text-slate-500 dark:text-slate-400';

export const TaskDataGrid = ({ task }: TaskDataGridProps) => {
  const tags = task.tags ?? [];

  return (
    <div className="grid gap-4 rounded-2xl border border-slate-200 p-5 md:grid-cols-2 dark:border-slate-700">
      <div className={infoFieldClassNames}>
        <Hash className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Code</p>
        {task.code ? (
          <TaskCodeBadge code={task.code} className="col-start-2 w-fit" />
        ) : (
          <p className={infoValueClassNames}>No code assigned.</p>
        )}
      </div>

      <div className={infoFieldClassNames}>
        <CircleCheck className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Status</p>
        <TaskStatusBadge status={task.status} label={task.project_column?.name} color={task.project_column?.color} />
      </div>

      <div className={infoFieldClassNames}>
        <ClockArrowUp className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Priority</p>
        <TaskPriorityBadge priority={task.priority} />
      </div>

      <div className={infoFieldClassNames}>
        <User className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Author</p>
        <p className={infoValueClassNames}>{task.author.name}</p>
      </div>

      <div className={infoFieldClassNames}>
        <User className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Responsible</p>
        <p className={infoValueClassNames}>{task.responsible?.name || '-'}</p>
      </div>

      <div className={infoFieldClassNames}>
        <Calendar className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Due date</p>
        <p className={infoValueClassNames}>{formatTaskDueDate(task.due_date)}</p>
      </div>

      <div className={infoFieldClassNames}>
        <Calendar className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Created at</p>
        <p className={infoValueClassNames}>{formatDateString(task.created_at)}</p>
      </div>

      <div className={infoFieldClassNames}>
        <Calendar className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Last updated</p>
        <p className={infoValueClassNames}>{formatDateString(task.updated_at)}</p>
      </div>

      <div className={cn(infoFieldClassNames, 'md:col-span-2')}>
        <Tag className={infoIconClassNames} />
        <p className={infoTitleClassNames}>Tags</p>
        <div className="col-span-2 flex w-full flex-wrap gap-2">
          {tags.length > 0 && (
            <>
              {tags.map((tag) => (
                <div
                  key={tag}
                  className="rounded-full border border-slate-200 px-2.5 py-1 text-xs font-medium text-slate-600 dark:border-slate-700 dark:text-slate-300"
                >
                  {tag}
                </div>
              ))}
            </>
          )}
          {tags.length === 0 && <p className={infoValueClassNames}>No tags added.</p>}
        </div>
      </div>
    </div>
  );
};

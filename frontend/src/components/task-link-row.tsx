import { ChevronRight } from 'lucide-react';
import { Link, useNavigate } from '@tanstack/react-router';
import type { Task } from '@/types/task';

interface TaskLinkRowProps {
  task: Pick<Task, 'id' | 'title' | 'project'>;
  trailingContent?: React.ReactNode;
}

export const TaskLinkRow = ({ task, trailingContent }: TaskLinkRowProps) => {
  const navigate = useNavigate();
  const projectId = task.project?.id ?? '';
  const openTask = () => {
    navigate({
      to: '/projects/$projectId',
      params: { projectId },
      search: { taskId: task.id },
    });
  };

  return (
    <div
      className="group flex w-full cursor-pointer items-center justify-between gap-4 border-b border-slate-200 pb-2 dark:border-slate-700"
      role="button"
      tabIndex={0}
      onClick={openTask}
      onKeyDown={(event) => {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          openTask();
        }
      }}
    >
      <div className="min-w-0 flex-1">
        <p className="truncate text-base font-medium text-slate-900 hover:underline dark:text-slate-100">
          {task.title}
        </p>

        <Link
          to="/projects/$projectId"
          params={{ projectId }}
          className="block w-fit max-w-full"
          onClick={(event) => {
            event.stopPropagation();
          }}
          onKeyDown={(event) => {
            event.stopPropagation();
          }}
        >
          <p className="truncate text-sm text-slate-500 hover:underline dark:text-slate-400">{task.project?.name}</p>
        </Link>
      </div>

      <div className="flex shrink-0 items-center gap-4">
        {trailingContent}
        <button
          type="button"
          aria-label={`Open task ${task.title}`}
          className="rounded-sm p-1 text-slate-500 transition-colors hover:text-slate-700 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none dark:text-slate-400 dark:hover:text-slate-300 dark:focus-visible:ring-slate-600"
          onClick={(event) => {
            event.stopPropagation();
            openTask();
          }}
        >
          <ChevronRight className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
};

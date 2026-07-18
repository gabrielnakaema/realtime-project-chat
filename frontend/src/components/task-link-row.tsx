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
      className="group border-border flex w-full cursor-pointer items-center justify-between gap-4 border-b pb-2"
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
        <p className="text-foreground truncate text-base font-medium hover:underline">{task.title}</p>

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
          <p className="text-muted-foreground truncate text-sm hover:underline">{task.project?.name}</p>
        </Link>
      </div>

      <div className="flex shrink-0 items-center gap-4">
        {trailingContent}
        <button
          type="button"
          aria-label={`Open task ${task.title}`}
          className="text-muted-foreground hover:text-foreground focus-visible:ring-ring rounded-sm p-1 transition-colors focus-visible:ring-2 focus-visible:outline-none"
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

import { Link } from '@tanstack/react-router';
import { differenceInCalendarDays, parseISO } from 'date-fns';
import { Loader2, RotateCcw } from 'lucide-react';
import type { Task } from '@/types/task';
import { useUserDueTasks } from '@/hooks/use-user-due-tasks';
import { cn } from '@/lib/utils';
import { Button } from '@/shared/components/button';

export const UpcomingDeadlines = () => {
  const { data, isError, isFetchingNextPage, isLoading, refetch, sentinelRef } = useUserDueTasks();

  return (
    <section aria-labelledby="upcoming-deadlines-heading">
      <h2 id="upcoming-deadlines-heading" className="text-foreground mb-3 text-[13px] font-semibold">
        Upcoming Deadlines
      </h2>

      <div className="border-border bg-card rounded-xl border px-4">
        {isLoading && <DeadlinesSkeleton />}

        {isError && !isLoading && (
          <PanelMessage
            title="Deadlines could not be loaded."
            action={
              <Button type="button" variant="ghost" size="sm" onClick={() => refetch()}>
                <RotateCcw className="size-3.5" />
                Try again
              </Button>
            }
          />
        )}

        {!isLoading && !isError && !data.length && (
          <PanelMessage title="Nothing due soon" description="Assigned tasks with due dates will appear here." />
        )}

        {!isLoading && !isError && !!data.length && (
          <div className="max-h-[192px] overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            {data.map((task) => (
              <DeadlineRow key={task.id} task={task} />
            ))}
            <div ref={sentinelRef} className="h-px" aria-hidden="true" />
            {isFetchingNextPage && (
              <div className="flex justify-center py-3">
                <Loader2 className="text-muted-foreground size-4 animate-spin" aria-label="Loading more deadlines" />
              </div>
            )}
          </div>
        )}
      </div>
    </section>
  );
};

const DeadlineRow = ({ task }: { task: Task }) => {
  const projectId = task.project?.id ?? task.project_id;
  const dueStatus = getDueStatus(task.due_date);

  return (
    <Link
      to="/projects/$projectId"
      params={{ projectId }}
      search={{ taskId: task.id }}
      className="border-border group flex items-start justify-between gap-3 border-b py-3 last:border-b-0"
    >
      <div className="min-w-0">
        <p className="text-foreground truncate text-[12.5px] leading-[1.4] group-hover:underline">{task.title}</p>
        <p className="text-primary mt-1 truncate text-[10.5px] font-medium">{task.project?.name ?? 'Project'}</p>
      </div>
      <span
        className={cn(
          'mt-px shrink-0 font-mono text-[11px] font-semibold whitespace-nowrap tabular-nums',
          dueStatus.tone === 'overdue' && 'text-destructive',
          dueStatus.tone === 'soon' && 'text-warning',
          dueStatus.tone === 'default' && 'text-muted-foreground',
        )}
      >
        {dueStatus.label}
      </span>
    </Link>
  );
};

export const getDueStatus = (dueDate: string | null, now = new Date()) => {
  if (!dueDate) {
    return { label: 'No due date', tone: 'default' as const };
  }

  const days = differenceInCalendarDays(parseISO(dueDate), now);
  if (days < 0) {
    return { label: `${Math.abs(days)}d overdue`, tone: 'overdue' as const };
  }
  if (days === 0) {
    return { label: 'Due today', tone: 'soon' as const };
  }
  if (days === 1) {
    return { label: 'Due tomorrow', tone: 'soon' as const };
  }
  return { label: `Due in ${days}d`, tone: 'default' as const };
};

const PanelMessage = ({
  action,
  description,
  title,
}: {
  action?: React.ReactNode;
  description?: string;
  title: string;
}) => (
  <div className="flex min-h-36 flex-col items-center justify-center py-6 text-center">
    <p className="text-foreground text-xs font-medium">{title}</p>
    {description && <p className="text-muted-foreground mt-1 text-[11px]">{description}</p>}
    {action && <div className="mt-3">{action}</div>}
  </div>
);

const DeadlinesSkeleton = () => (
  <div className="animate-pulse" aria-label="Loading upcoming deadlines">
    {Array.from({ length: 4 }).map((_, index) => (
      <div key={index} className="border-border flex items-start justify-between gap-3 border-b py-3 last:border-b-0">
        <div className="flex-1 space-y-2">
          <div className="bg-muted h-3 w-3/4 rounded" />
          <div className="bg-muted h-2.5 w-1/3 rounded" />
        </div>
        <div className="bg-muted h-3 w-16 rounded" />
      </div>
    ))}
  </div>
);

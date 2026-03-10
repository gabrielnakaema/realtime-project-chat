import { ChevronRight, Loader2 } from 'lucide-react';
import { Link, useNavigate } from '@tanstack/react-router';
import { differenceInCalendarDays, parseISO } from 'date-fns';
import { useUserDueTasks } from '@/hooks/use-user-due-tasks';
import { cn } from '@/lib/utils';

const BADGE_BASE = 'text-xs font-medium rounded-md px-2 py-0.5';

function formatDueDate(date: string | null): string {
  if (!date) return '-';
  return new Date(date).toLocaleDateString([], { month: 'short', day: 'numeric' });
}

function DueBadge({ date }: { date: string | null }) {
  if (!date) return null;
  const days = differenceInCalendarDays(parseISO(date), new Date());
  if (days < 0) {
    return (
      <span className={cn(BADGE_BASE, 'text-white dark:text-white', 'bg-red-900')}>{Math.abs(days)} days ago</span>
    );
  }
  if (days === 0) {
    return <span className={cn(BADGE_BASE, 'text-white dark:text-white', 'bg-blue-800')}>Today</span>;
  }
  if (days === 1) {
    return <span className={cn(BADGE_BASE, 'text-white dark:text-white', 'bg-green-800')}>Tomorrow</span>;
  }
  return <span className={cn(BADGE_BASE, 'text-white dark:text-white', 'bg-slate-700')}>{days} days</span>;
}

export const UserDueTasks = () => {
  const { data, isLoading, isFetchingNextPage, sentinelRef } = useUserDueTasks();
  const navigate = useNavigate();

  if (!isLoading && !data.length) return null;
  if (isLoading) return <UserDueTasksSkeleton />;

  return (
    <section className="flex flex-col gap-6 rounded-lg border border-slate-200 bg-white p-6 transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800">
      <div className="flex flex-col gap-1">
        <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Upcoming deadlines</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400">Tasks assigned to you with approaching due dates</p>
      </div>

      <div className="flex max-h-96 flex-col gap-4 overflow-y-auto pr-1">
        {data.map((task) => (
          <article
            key={task.id}
            className="group flex w-full cursor-pointer flex-row items-center justify-between gap-2 border-b border-slate-200 pb-2 dark:border-slate-700"
            tabIndex={0}
            role="button"
            onClick={() => {
              navigate({
                to: '/projects/$projectId',
                params: { projectId: task.project?.id ?? '' },
                search: { taskId: task.id },
              });
            }}
          >
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{task.title}</p>
              <Link
                to="/projects/$projectId"
                params={{ projectId: task.project?.id ?? '' }}
                onClick={(e) => {
                  e.stopPropagation();
                }}
              >
                <p className="text-sm text-slate-500 hover:underline dark:text-slate-400">{task.project?.name}</p>
              </Link>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500 dark:text-slate-400">{formatDueDate(task.due_date)}</span>
              <DueBadge date={task.due_date} />

              <ChevronRight className="invisible h-4 w-4 text-slate-500 opacity-0 transition-all group-hover:visible group-hover:text-slate-700 group-hover:opacity-100 hover:text-slate-700 dark:text-slate-400 dark:group-hover:text-slate-300 dark:hover:text-slate-300" />
            </div>
          </article>
        ))}
        {isFetchingNextPage && (
          <div className="flex items-center justify-center">
            <Loader2 className="h-4 w-4 animate-spin" />
          </div>
        )}
        <div ref={sentinelRef} className="h-1" aria-hidden="true" />
      </div>
    </section>
  );
};

const UserDueTasksSkeleton = () => (
  <section className="flex flex-col gap-6 rounded-lg border border-slate-200 bg-white p-6 transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800">
    <div className="flex flex-col gap-1">
      <div className="h-6 w-48 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
      <div className="h-4 w-64 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
    </div>

    <div className="flex max-h-96 flex-col gap-4 overflow-y-auto pr-1">
      {Array.from({ length: 5 }).map((_, i) => (
        <article
          key={i}
          className="group flex w-full flex-row items-center justify-between gap-2 border-b border-slate-200 pb-2 dark:border-slate-700"
        >
          <div className="flex flex-col gap-1">
            <div className="h-4 w-48 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
            <div className="h-3 w-24 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
          </div>
          <div className="h-4 w-20 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
        </article>
      ))}
    </div>
  </section>
);

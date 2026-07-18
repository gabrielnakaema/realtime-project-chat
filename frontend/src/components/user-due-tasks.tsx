import { Loader2 } from 'lucide-react';
import { differenceInCalendarDays, parseISO } from 'date-fns';
import { useUserDueTasks } from '@/hooks/use-user-due-tasks';
import { TaskLinkRow } from '@/components/task-link-row';
import { cn } from '@/lib/utils';
import { formatMonthDay } from '@/utils/date';

const BADGE_BASE = 'text-xs font-medium rounded-md px-2 py-0.5';

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

  if (!isLoading && !data.length) return null;
  if (isLoading) return <UserDueTasksSkeleton />;

  return (
    <section className="border-border bg-card flex flex-col gap-6 rounded-lg border p-6 transition-shadow hover:shadow-lg">
      <div className="flex flex-col gap-1">
        <h2 className="text-foreground text-xl font-semibold">Upcoming deadlines</h2>
        <p className="text-muted-foreground text-sm">Tasks assigned to you with approaching due dates</p>
      </div>

      <div className="flex max-h-96 flex-col gap-4 overflow-y-auto pr-1">
        {data.map((task) => (
          <TaskLinkRow
            key={task.id}
            task={task}
            trailingContent={
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground text-sm">{formatMonthDay(task.due_date)}</span>
                <DueBadge date={task.due_date} />
              </div>
            }
          />
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
  <section className="border-border bg-card flex flex-col gap-6 rounded-lg border p-6 transition-shadow hover:shadow-lg">
    <div className="flex flex-col gap-1">
      <div className="bg-muted h-6 w-48 animate-pulse rounded" />
      <div className="bg-muted h-4 w-64 animate-pulse rounded" />
    </div>

    <div className="flex max-h-96 flex-col gap-4 overflow-y-auto pr-1">
      {Array.from({ length: 5 }).map((_, i) => (
        <article
          key={i}
          className="group border-border flex w-full flex-row items-center justify-between gap-2 border-b pb-2"
        >
          <div className="flex flex-col gap-1">
            <div className="bg-muted h-4 w-48 animate-pulse rounded" />
            <div className="bg-muted h-3 w-24 animate-pulse rounded" />
          </div>
          <div className="bg-muted h-4 w-20 animate-pulse rounded" />
        </article>
      ))}
    </div>
  </section>
);

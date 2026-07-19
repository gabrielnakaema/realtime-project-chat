import { Link } from '@tanstack/react-router';
import { Loader2, RotateCcw } from 'lucide-react';
import { ProjectActivityText } from '@/features/projects/components/project-activity-text';
import { useUserProjectActivities } from '@/features/projects/hooks/use-user-project-activities';
import { Button } from '@/shared/components/button';
import { formatRelativeActivityDateString } from '@/shared/utils/format-relative-activity';

export const RecentProjectActivity = () => {
  const { data, queryResult, sentinelRef } = useUserProjectActivities();
  const { isError, isFetchingNextPage, isLoading, refetch } = queryResult;

  return (
    <section aria-labelledby="recent-activity-heading">
      <h2 id="recent-activity-heading" className="text-foreground mb-3 text-[13px] font-semibold">
        Recent Activity
      </h2>

      <div className="border-border bg-card rounded-xl border px-4">
        {isLoading && <ActivitySkeleton />}

        {isError && (
          <PanelMessage
            title="Activity could not be loaded."
            action={
              <Button type="button" variant="ghost" size="sm" onClick={() => refetch()}>
                <RotateCcw className="size-3.5" />
                Try again
              </Button>
            }
          />
        )}

        {!isLoading && !isError && !data.length && (
          <PanelMessage title="No recent activity" description="Project updates will appear here." />
        )}

        {!isLoading && !isError && !!data.length && (
          <div className="max-h-[315px] overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            {data.map((activity, index) => (
              <article
                key={activity.id}
                className="border-border flex items-start gap-2.5 border-b py-3 last:border-b-0"
              >
                <span
                  className={
                    index < 2 ? 'bg-primary mt-1.5 size-1.5 shrink-0 rounded-full' : 'mt-1.5 size-1.5 shrink-0'
                  }
                  aria-hidden="true"
                />
                <div className="min-w-0 flex-1 [&_p]:text-[12.5px] [&_p]:leading-[1.45]">
                  <ProjectActivityText activity={activity} />
                  <div className="mt-1 flex flex-wrap items-center gap-2">
                    {activity.project && (
                      <Link
                        to="/projects/$projectId"
                        params={{ projectId: activity.project.id }}
                        className="text-primary text-[10.5px] font-medium hover:underline"
                      >
                        {activity.project.name}
                      </Link>
                    )}
                    <time
                      dateTime={activity.created_at}
                      className="text-muted-foreground font-mono text-[10.5px] tabular-nums"
                    >
                      {formatRelativeActivityDateString(activity.created_at)}
                    </time>
                  </div>
                </div>
              </article>
            ))}
            <div ref={sentinelRef} className="h-px" aria-hidden="true" />
            {isFetchingNextPage && (
              <div className="flex justify-center py-3">
                <Loader2 className="text-muted-foreground size-4 animate-spin" aria-label="Loading more activity" />
              </div>
            )}
          </div>
        )}
      </div>
    </section>
  );
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

const ActivitySkeleton = () => (
  <div className="animate-pulse" aria-label="Loading recent activity">
    {Array.from({ length: 4 }).map((_, index) => (
      <div key={index} className="border-border flex items-start gap-2.5 border-b py-3 last:border-b-0">
        <div className="bg-muted mt-1.5 size-1.5 rounded-full" />
        <div className="flex-1 space-y-2">
          <div className="bg-muted h-3 w-4/5 rounded" />
          <div className="bg-muted h-2.5 w-1/3 rounded" />
        </div>
      </div>
    ))}
  </div>
);

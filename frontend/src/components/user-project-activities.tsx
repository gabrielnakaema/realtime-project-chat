import { Loader2Icon } from 'lucide-react';
import { ProjectActivity } from './project-activity';
import { useUserProjectActivities } from '@/hooks/use-user-project-activities';

export const UserProjectActivities = () => {
  const { data, queryResult, sentinelRef } = useUserProjectActivities();
  const { isLoading, isFetchingNextPage } = queryResult;

  if (isLoading || !data.length) {
    return null;
  }

  return (
    <section className="border-border bg-card flex flex-col gap-6 rounded-lg border p-6 transition-shadow hover:shadow-lg">
      <div className="flex flex-col gap-1">
        <h2 className="text-foreground text-xl font-semibold">Recent activity</h2>
        <p className="text-muted-foreground text-sm">Latest updates from projects you're a member of</p>
      </div>

      <div className="flex max-h-96 flex-col gap-4 overflow-y-auto pr-1">
        {data.map((activity) => (
          <ProjectActivity key={activity.id} activity={activity} />
        ))}
        <div ref={sentinelRef} className="h-1" aria-hidden="true" />
        {isFetchingNextPage && (
          <div className="flex justify-center py-2">
            <Loader2Icon className="text-muted-foreground h-4 w-4 animate-spin" />
          </div>
        )}
      </div>
    </section>
  );
};

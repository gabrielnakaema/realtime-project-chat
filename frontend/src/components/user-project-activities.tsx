import { ProjectActivity } from './project-activity';
import { useUserProjectActivities } from '@/hooks/use-user-project-activities';

export const UserProjectActivities = () => {
  const { data, queryResult } = useUserProjectActivities();
  const { isLoading } = queryResult;

  if (isLoading || !data.length) {
    return null;
  }

  return (
    <section className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition-shadow p-6 flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Recent activity</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400">Latest updates from projects you're a member of</p>
      </div>

      <div className="flex flex-col gap-4">
        {data.map((activity) => (
          <ProjectActivity key={activity.id} activity={activity} />
        ))}
      </div>
    </section>
  );
};

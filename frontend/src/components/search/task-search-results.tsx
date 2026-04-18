import { ListTodo, Loader2 } from 'lucide-react';
import { TaskStatusBadge } from '@/components/task-status-badge';
import { TaskLinkRow } from '@/components/task-link-row';
import { SearchEmptyState } from '@/components/search/empty-state';
import { useSearchTasks } from '@/hooks/use-search-tasks';

interface TaskSearchResultsProps {
  query?: string;
}

export const TaskSearchResults = ({ query }: TaskSearchResultsProps) => {
  const { data: tasks, isLoading, isFetchingNextPage, sentinelRef } = useSearchTasks(query);
  const areTasksEmpty = !isLoading && !tasks.length;

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">Tasks</h2>
      {areTasksEmpty ? (
        <SearchEmptyState
          icon={<ListTodo className="h-4 w-4" />}
          title="No tasks found"
          description="There are no tasks for your search. Try adjusting your search terms."
        />
      ) : (
        <div className="flex w-full flex-col gap-4 overflow-hidden rounded-lg border border-slate-200 bg-white p-6 pr-2 transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800">
          <div className="flex max-h-96 w-full flex-col gap-4 overflow-y-auto pr-2">
            {isLoading && <TaskSearchSkeleton />}

            {tasks.map((task) => (
              <TaskLinkRow key={task.id} task={task} trailingContent={<TaskStatusBadge status={task.status} />} />
            ))}

            {isFetchingNextPage && (
              <div className="flex items-center justify-center">
                <Loader2 className="h-4 w-4 animate-spin" />
              </div>
            )}

            <div ref={sentinelRef} className="h-1" aria-hidden="true" />
          </div>
        </div>
      )}
    </section>
  );
};

const TaskSearchSkeleton = () => {
  return Array.from({ length: 5 }).map((_, index) => (
    <div
      key={index}
      className="flex w-full items-center justify-between rounded-md border-b pb-2 dark:border-slate-700"
    >
      <div className="flex flex-col gap-2">
        <div className="h-4 w-48 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
        <div className="h-3 w-72 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
      </div>
      <div className="h-4 w-20 animate-pulse rounded bg-slate-200 dark:bg-slate-600" />
    </div>
  ));
};

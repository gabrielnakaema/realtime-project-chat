import { Link, createFileRoute } from '@tanstack/react-router';
import { ChevronRight, ListTodo, Loader2, Package } from 'lucide-react';
import { z } from 'zod';
import { HeaderUser } from '@/components/header-user';
import { SearchBar } from '@/components/search-bar';
import { TaskStatusBadge } from '@/components/task-status-badge';
import { useSearchProjects } from '@/hooks/use-search-projects';
import { useSearchTasks } from '@/hooks/use-search-tasks';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';
import { sanitizeHTML } from '@/utils/html';

export const Route = createFileRoute('/search/')({
  component: RouteComponent,
  validateSearch: z.object({
    query: z.string().optional(),
  }),
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const { query } = Route.useSearch();

  const { data: projects, isLoading: isLoadingProjects } = useSearchProjects(query);

  const {
    data: tasks,
    isLoading: isLoadingTasks,
    isFetchingNextPage: isFetchingNextPageTasks,
    sentinelRef: sentinelRefTasks,
  } = useSearchTasks(query);

  const areTasksEmpty = !isLoadingTasks && !tasks.length;
  const areProjectsEmpty = !isLoadingProjects && !projects?.length;

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">TaskFlow</h1>
            </div>

            <SearchBar action="/search" searchName="query" formClassName="w-full max-w-md" initialValue={query} />
            <div className="flex items-center gap-4">
              <HeaderUser />
            </div>
          </div>
        </div>
      </header>
      <div className="flex w-full flex-col gap-8 px-6 py-12">
        <p className="text-lg font-medium text-slate-900 dark:text-slate-400">
          Showing results for <span className="text-slate-500 dark:text-slate-100">&quot;{query}&quot;</span>
        </p>
        <section className="flex flex-col gap-4">
          <h2 className="text-xl font-bold text-slate-900 dark:border-slate-700 dark:text-slate-100">Projects</h2>
          <div className="flex w-full max-w-full gap-4 overflow-x-auto overflow-y-auto">
            {isLoadingProjects && <ProjectSearchSkeleton />}
            {areProjectsEmpty && (
              <EmptyState
                icon={<Package className="h-4 w-4" />}
                title="No projects found"
                description="There are no projects for your search. Try adjusting your search terms."
              />
            )}
            {projects?.map((project) => (
              <Link
                key={project.id}
                to="/projects/$projectId"
                params={{ projectId: project.id }}
                className="w-full max-w-sm min-w-sm cursor-pointer rounded-lg border border-slate-200 bg-white transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800"
              >
                <div className="flex h-full flex-col gap-3 p-6">
                  <div className="flex items-center justify-between">
                    <h3 className="overflow-hidden text-lg font-semibold text-ellipsis whitespace-nowrap text-slate-900 dark:text-slate-100">
                      {project.name}
                    </h3>
                    <span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-1 text-xs font-medium text-slate-700 dark:bg-slate-700 dark:text-slate-300">
                      {project.members.length || 0} {project.members.length === 1 ? 'member' : 'members'}
                    </span>
                  </div>

                  <p className="line-clamp-3 max-h-16 min-h-16 overflow-hidden text-sm text-ellipsis text-slate-600 dark:text-slate-400">
                    <div dangerouslySetInnerHTML={{ __html: sanitizeHTML(project.description) }} />
                  </p>
                  <p className="mt-auto text-xs text-slate-500 dark:border-slate-700 dark:text-slate-400">
                    Last activity:{' '}
                    <time dateTime={project.updated_at}>
                      {project.updated_at ? formatRelativeActivityDateString(project.updated_at) : 'N/A'}
                    </time>
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </section>

        <section className="flex flex-col gap-4">
          <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">Tasks</h2>
          {areTasksEmpty ? (
            <EmptyState
              icon={<ListTodo className="h-4 w-4" />}
              title="No tasks found"
              description="There are no tasks for your search. Try adjusting your search terms."
            />
          ) : (
            <div className="flex w-full flex-col gap-4 overflow-hidden rounded-lg border border-slate-200 bg-white p-6 pr-2 transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800">
              <div className="flex max-h-96 w-full flex-col gap-4 overflow-y-auto pr-2">
                {isLoadingTasks && <TaskSearchSkeleton />}

                {tasks.map((task) => (
                  <article
                    key={task.id}
                    className="group flex w-full cursor-pointer items-center justify-between border-b border-slate-200 pb-2 dark:border-slate-700"
                    role="button"
                    tabIndex={0}
                    onClick={() =>
                      navigate({
                        to: '/projects/$projectId',
                        params: { projectId: task.project?.id ?? '' },
                        search: { taskId: task.id },
                      })
                    }
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        navigate({
                          to: '/projects/$projectId',
                          params: { projectId: task.project?.id ?? '' },
                          search: { taskId: task.id },
                        });
                      }
                    }}
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex flex-col gap-0.5">
                        <p className="text-base font-medium text-slate-900 hover:underline dark:text-slate-100">
                          {task.title}
                        </p>
                        <Link to="/projects/$projectId" params={{ projectId: task.project?.id ?? '' }}>
                          <p className="text-sm text-slate-500 hover:underline dark:text-slate-400">
                            {task.project?.name}
                          </p>
                        </Link>
                      </div>
                    </div>

                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-1">
                        <TaskStatusBadge status={task.status} />
                      </div>

                      <ChevronRight className="invisible h-4 w-4 text-slate-500 opacity-0 transition-opacity group-hover:visible group-hover:text-slate-700 group-hover:opacity-100 dark:text-slate-400 dark:group-hover:text-slate-300" />
                    </div>
                  </article>
                ))}
                {isFetchingNextPageTasks && (
                  <div className="flex items-center justify-center">
                    <Loader2 className="h-4 w-4 animate-spin" />
                  </div>
                )}
                <div ref={sentinelRefTasks} className="h-1" aria-hidden="true" />
              </div>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
const ProjectSearchSkeleton = () => {
  return Array.from({ length: 5 }).map((_, index) => (
    <div
      key={index}
      className="w-full min-w-sm cursor-pointer rounded-lg border border-slate-200 bg-white transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800"
    >
      <div className="flex flex-col gap-4 p-6">
        <div className="mb-4 flex items-center justify-between">
          <div className="h-4 w-1/4 animate-pulse rounded-full bg-slate-100 dark:bg-slate-700" />
          <div className="h-4 w-1/5 animate-pulse rounded-full bg-slate-100 dark:bg-slate-700" />
        </div>

        <div className="space-y-1">
          <div className="h-4 w-full animate-pulse rounded-full bg-slate-100 dark:bg-slate-700" />
          <div className="h-4 w-full animate-pulse rounded-full bg-slate-100 dark:bg-slate-700" />
          <div className="h-4 w-4/5 animate-pulse rounded-full bg-slate-100 dark:bg-slate-700" />
        </div>
        <div className="h-4 w-1/3 animate-pulse rounded-full bg-slate-100 dark:bg-slate-700" />
      </div>
    </div>
  ));
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

interface EmptyStateProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}

const EmptyState = ({ icon, title, description }: EmptyStateProps) => {
  return (
    <div className="border-border flex w-full flex-col items-center justify-center rounded-lg border border-dashed px-6 py-12 text-center">
      <div className="bg-secondary text-muted-foreground mb-4 flex h-12 w-12 items-center justify-center rounded-full">
        {icon}
      </div>
      <h3 className="text-foreground mb-1 text-sm font-medium">{title}</h3>
      <p className="text-muted-foreground mb-4 max-w-sm text-sm">{description}</p>
    </div>
  );
};

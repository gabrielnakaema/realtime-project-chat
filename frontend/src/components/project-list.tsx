import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { Link } from '@tanstack/react-router';
import { CreateProject } from './create-project';
import { listProjects } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';

export const ProjectList = () => {
  const { data: projects, isLoading } = useQuery({
    queryKey: projectQueryKeys.list,
    queryFn: listProjects,
  });

  if (isLoading) {
    return <ProjectListSkeleton />;
  }

  return (
    <section className="flex w-full flex-col gap-6">
      <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Your Projects</h2>
      {!projects?.length && (
        <div className="rounded-lg border border-slate-200 bg-white p-12 text-center dark:border-slate-700 dark:bg-slate-800">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800">
            <Plus className="h-8 w-8 text-slate-400" />
          </div>
          <h3 className="mb-2 text-lg font-semibold text-slate-900 dark:text-slate-100">No projects yet</h3>
          <p className="mb-6 text-slate-600 dark:text-slate-400">
            Create your first project to start organizing tasks and collaborating with your team.
          </p>
          <div className="flex justify-center">
            <CreateProject />
          </div>
        </div>
      )}

      {!!projects?.length && (
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <Link
              key={project.id}
              to="/projects/$projectId"
              params={{ projectId: project.id }}
              className="cursor-pointer rounded-lg border border-slate-200 bg-white transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800"
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

                <p className="line-clamp-3 min-h-16 text-sm text-slate-600 dark:text-slate-400">
                  {project.description}
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
      )}
    </section>
  );
};

export const ProjectListSkeleton = () => {
  return (
    <section className="flex w-full flex-col gap-6">
      <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Your Projects</h2>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <div
            key={index}
            className="cursor-pointer rounded-lg border border-slate-200 bg-white transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800"
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
        ))}
      </div>
    </section>
  );
};

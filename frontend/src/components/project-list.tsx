import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { CreateProject } from './project-form/create-project';
import { ProjectCard, ProjectCardSkeleton } from './project-card';
import { listProjects } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';

export const ProjectList = () => {
  const { data: projects, isLoading } = useQuery({
    queryKey: projectQueryKeys.list,
    queryFn: () => listProjects(),
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
        <div className="flex w-full max-w-full items-stretch gap-4 overflow-x-auto pb-3">
          {projects.map((project) => (
            <ProjectCard key={project.id} project={project} />
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
          <ProjectCardSkeleton
            key={index}
            className="rounded-lg border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800"
          />
        ))}
      </div>
    </section>
  );
};

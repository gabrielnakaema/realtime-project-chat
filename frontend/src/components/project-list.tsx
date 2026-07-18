import { Plus } from 'lucide-react';
import { ProjectCard, ProjectCardSkeleton } from './project-card';
import { CreateProject } from './project-form/create-project';
import { useListProjects } from '@/hooks/use-list-projects';

export const ProjectList = () => {
  const { data: projects, isLoading } = useListProjects();

  if (isLoading) {
    return <ProjectListSkeleton />;
  }

  return (
    <section className="flex w-full flex-col gap-6">
      <h2 className="text-foreground text-xl font-semibold">Your Projects</h2>
      {!projects?.length && (
        <div className="border-border bg-card rounded-lg border p-12 text-center">
          <div className="bg-muted mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
            <Plus className="text-muted-foreground h-8 w-8" />
          </div>
          <h3 className="text-foreground mb-2 text-lg font-semibold">No projects yet</h3>
          <p className="text-muted-foreground mb-6">
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
      <h2 className="text-foreground text-xl font-semibold">Your Projects</h2>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <ProjectCardSkeleton key={index} className="border-border bg-card rounded-lg border" />
        ))}
      </div>
    </section>
  );
};

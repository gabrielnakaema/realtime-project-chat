import { Link } from '@tanstack/react-router';
import { Plus, RotateCcw } from 'lucide-react';
import type { Project } from '@/features/projects/types/project';
import { ProjectCardSkeleton } from '@/features/projects/components/project-card';
import { CreateProject } from '@/features/projects/components/project-form/create-project';
import { useListProjects } from '@/features/projects/hooks/use-list-projects';
import { cn } from '@/lib/utils';
import { Button } from '@/shared/components/button';
import { formatRelativeActivityDateString } from '@/shared/utils/format-relative-activity';
import { richTextListClassName, sanitizeHTML } from '@/shared/utils/html';

export const ProjectsGrid = () => {
  const { data: projects, isError, isLoading, refetch } = useListProjects();

  return (
    <section aria-labelledby="projects-heading">
      <div className="mb-5 flex items-center justify-between gap-4">
        <h1 id="projects-heading" className="text-foreground text-[19px] font-semibold tracking-[-0.01em]">
          Projects
        </h1>

        <CreateProject
          trigger={
            <Button type="button" variant="outline" size="sm" className="border-primary/50 bg-primary/10 text-primary">
              <Plus className="size-3.5" />
              New project
            </Button>
          }
        />
      </div>

      {isLoading && <ProjectsGridSkeleton />}

      {isError && (
        <div className="border-border bg-card flex min-h-48 flex-col items-center justify-center rounded-xl border p-8 text-center">
          <p className="text-foreground text-sm font-medium">Projects could not be loaded.</p>
          <p className="text-muted-foreground mt-1 text-xs">Check your connection and try again.</p>
          <Button type="button" variant="outline" size="sm" className="mt-4" onClick={() => refetch()}>
            <RotateCcw className="size-3.5" />
            Try again
          </Button>
        </div>
      )}

      {!isLoading && !isError && !projects?.length && <EmptyProjects />}

      {!isLoading && !isError && !!projects?.length && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <ProjectSummaryCard key={project.id} project={project} />
          ))}
        </div>
      )}
    </section>
  );
};

export const ProjectSummaryCard = ({ project }: { project: Project }) => {
  const memberCount = project.members.length;
  const columnCount = project.columns.length;

  return (
    <Link
      to="/projects/$projectId"
      params={{ projectId: project.id }}
      className="border-border bg-card hover:border-primary/40 focus-visible:ring-ring group flex min-h-48 flex-col rounded-xl border p-5 transition-colors focus-visible:ring-2 focus-visible:outline-none"
    >
      <h2 className="text-foreground truncate text-sm font-semibold">{project.name}</h2>
      <div
        className={cn(
          'text-muted-foreground mt-2 line-clamp-2 min-h-10 overflow-hidden text-xs leading-5',
          richTextListClassName,
        )}
        dangerouslySetInnerHTML={{ __html: sanitizeHTML(project.description || 'No description yet.') }}
      />

      <dl className="mt-4 flex items-center gap-5">
        <div className="flex items-baseline gap-1">
          <dt className="text-muted-foreground order-2 text-xs font-normal">members</dt>
          <dd className="text-foreground order-1 font-mono text-sm font-semibold tabular-nums">{memberCount}</dd>
        </div>
        <div className="flex items-baseline gap-1">
          <dt className="text-muted-foreground order-2 text-xs font-normal">columns</dt>
          <dd className="text-foreground order-1 font-mono text-sm font-semibold tabular-nums">{columnCount}</dd>
        </div>
      </dl>

      <div className="border-border mt-auto flex items-center justify-between gap-3 border-t pt-3">
        <ProjectMemberAvatars project={project} />
        <time dateTime={project.updated_at} className="text-muted-foreground truncate font-mono text-xs tabular-nums">
          {project.updated_at ? formatRelativeActivityDateString(project.updated_at) : 'No activity'}
        </time>
      </div>
    </Link>
  );
};

const ProjectMemberAvatars = ({ project }: { project: Project }) => {
  const visibleMembers = project.members.slice(0, 4);
  const remaining = project.members.length - visibleMembers.length;

  if (!visibleMembers.length) {
    return <span className="text-muted-foreground text-[11px]">No members</span>;
  }

  return (
    <div className="flex -space-x-1.5" aria-label={`${project.members.length} project members`}>
      {visibleMembers.map((member) => (
        <span
          key={member.id}
          title={member.user.name}
          className="border-card bg-muted text-muted-foreground flex size-6 items-center justify-center rounded-full border-2 text-xs font-semibold"
        >
          {getInitials(member.user.name)}
        </span>
      ))}
      {remaining > 0 && (
        <span className="border-card bg-muted text-muted-foreground flex size-6 items-center justify-center rounded-full border-2 text-xs font-semibold">
          +{remaining}
        </span>
      )}
    </div>
  );
};

const getInitials = (name: string) => {
  const words = name.trim().split(/\s+/).filter(Boolean);
  return words
    .slice(0, 2)
    .map((word) => word.charAt(0).toUpperCase())
    .join('');
};

const EmptyProjects = () => (
  <div className="flex flex-col items-center justify-center px-6 py-16 text-center sm:py-20">
    <div className="relative mb-8 h-36 w-56" aria-hidden="true">
      <div className="border-border bg-card absolute top-6 left-0 h-24 w-16 rounded-xl border" />
      <div className="border-border bg-card absolute top-6 right-0 h-24 w-16 rounded-xl border" />
      <div className="border-border bg-card absolute top-0 left-20 h-36 w-16 rounded-xl border shadow-xl">
        <div className="bg-muted mx-3 mt-4 h-2 w-10 rounded" />
        <div className="bg-muted mx-3 mt-2 h-2 w-7 rounded" />
        <div className="border-primary/50 text-primary mx-3 mt-5 flex h-8 items-center justify-center rounded-lg border border-dashed">
          <Plus className="size-4" />
        </div>
      </div>
    </div>
    <h2 className="text-foreground text-lg font-semibold">Create your first project</h2>
    <p className="text-muted-foreground mt-2 max-w-sm text-sm leading-6">
      Set up a board, organize tasks, and invite collaborators to start coordinating work in real time.
    </p>
    <div className="mt-6">
      <CreateProject
        trigger={
          <Button type="button" size="sm">
            <Plus className="size-4" />
            New project
          </Button>
        }
      />
    </div>
  </div>
);

const ProjectsGridSkeleton = () => (
  <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3" aria-label="Loading projects">
    {Array.from({ length: 3 }).map((_, index) => (
      <ProjectCardSkeleton key={index} className="border-border bg-card min-h-48 rounded-xl border" />
    ))}
  </div>
);

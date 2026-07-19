import { Link } from '@tanstack/react-router';
import type { Project } from '@/features/projects/types/project';
import { formatRelativeActivityDateString } from '@/shared/utils/format-relative-activity';
import { richTextListClassName, sanitizeHTML } from '@/shared/utils/html';
import { cn } from '@/lib/utils';

interface ProjectCardProps {
  project: Project;
  className?: string;
}

export const ProjectCard = ({ project, className }: ProjectCardProps) => {
  return (
    <Link
      to="/projects/$projectId"
      params={{ projectId: project.id }}
      className={
        className ??
        'border-border bg-card w-full max-w-sm min-w-sm cursor-pointer rounded-lg border transition-shadow hover:shadow-lg'
      }
    >
      <div className="flex h-full flex-col gap-3 p-6">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-foreground overflow-hidden text-lg font-semibold text-ellipsis whitespace-nowrap">
            {project.name}
          </h3>
          <span className="bg-muted text-foreground inline-flex items-center rounded-full px-2 py-1 text-xs font-medium">
            {project.members.length || 0} {project.members.length === 1 ? 'member' : 'members'}
          </span>
        </div>

        <ProjectCardDescription description={project.description} />

        <p className="text-muted-foreground mt-auto text-xs">
          Last activity:{' '}
          <time dateTime={project.updated_at}>
            {project.updated_at ? formatRelativeActivityDateString(project.updated_at) : 'N/A'}
          </time>
        </p>
      </div>
    </Link>
  );
};

export const ProjectCardSkeleton = ({ className }: { className?: string }) => {
  return (
    <div
      className={
        className ??
        'border-border bg-card w-full min-w-sm cursor-pointer rounded-lg border transition-shadow hover:shadow-lg'
      }
    >
      <div className="flex flex-col gap-4 p-6">
        <div className="mb-4 flex items-center justify-between">
          <div className="bg-muted h-4 w-1/4 animate-pulse rounded-full" />
          <div className="bg-muted h-4 w-1/5 animate-pulse rounded-full" />
        </div>

        <div className="space-y-1">
          <div className="bg-muted h-4 w-full animate-pulse rounded-full" />
          <div className="bg-muted h-4 w-full animate-pulse rounded-full" />
          <div className="bg-muted h-4 w-4/5 animate-pulse rounded-full" />
        </div>
        <div className="bg-muted h-4 w-1/3 animate-pulse rounded-full" />
      </div>
    </div>
  );
};

const ProjectCardDescription = ({ description }: { description: string }) => {
  return (
    <div
      className={cn(
        'text-muted-foreground line-clamp-3 max-h-16 min-h-16 overflow-hidden text-sm text-ellipsis',
        richTextListClassName,
      )}
      dangerouslySetInnerHTML={{ __html: sanitizeHTML(description) }}
    />
  );
};

import { Link } from '@tanstack/react-router';
import type { Project } from '@/types/project';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';
import { sanitizeHTML } from '@/utils/html';

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
        'w-full max-w-sm min-w-sm cursor-pointer rounded-lg border border-slate-200 bg-white transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800'
      }
    >
      <div className="flex h-full flex-col gap-3 p-6">
        <div className="flex items-center justify-between gap-3">
          <h3 className="overflow-hidden text-lg font-semibold text-ellipsis whitespace-nowrap text-slate-900 dark:text-slate-100">
            {project.name}
          </h3>
          <span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-1 text-xs font-medium text-slate-700 dark:bg-slate-700 dark:text-slate-300">
            {project.members.length || 0} {project.members.length === 1 ? 'member' : 'members'}
          </span>
        </div>

        <ProjectCardDescription description={project.description} />

        <p className="mt-auto text-xs text-slate-500 dark:border-slate-700 dark:text-slate-400">
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
        'w-full min-w-sm cursor-pointer rounded-lg border border-slate-200 bg-white transition-shadow hover:shadow-lg dark:border-slate-700 dark:bg-slate-800'
      }
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
  );
};

const ProjectCardDescription = ({ description }: { description: string }) => {
  return (
    <div
      className="line-clamp-3 max-h-16 min-h-16 overflow-hidden text-sm text-ellipsis text-slate-600 dark:text-slate-400"
      dangerouslySetInnerHTML={{ __html: sanitizeHTML(description) }}
    />
  );
};

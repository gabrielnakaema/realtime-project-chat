import { Link } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';
import { ProjectHeader } from '../project-header';
import { useProjectDetails } from '@/hooks/use-project-details';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/shared/components/button';

export const ProjectSettingsHeader = ({ projectId }: { projectId: string }) => {
  const { data: project } = useProjectDetails(projectId);

  if (!project) {
    return null;
  }

  return (
    <ProjectHeader>
      <div className="flex w-full items-center justify-between">
        <div className="flex items-center gap-2">
          <ProjectHeader.Logo />
          <div className="flex items-center gap-2 pl-2">
            <Link
              to="/projects/$projectId"
              params={{
                projectId,
              }}
              className={cn(buttonVariants({ variant: 'link', size: 'sm' }), 'text-muted-foreground text-sm')}
            >
              {project.name}
            </Link>
            <span className="text-muted-foreground text-sm">/</span>
            <p className="pl-2 text-sm font-semibold">Settings</p>
          </div>
        </div>

        <Link
          to="/projects/$projectId"
          params={{ projectId }}
          className={cn(buttonVariants({ variant: 'ghost', size: 'sm' }))}
        >
          <ArrowLeft className="h-4 w-4" />
          Back to board
        </Link>
      </div>
    </ProjectHeader>
  );
};

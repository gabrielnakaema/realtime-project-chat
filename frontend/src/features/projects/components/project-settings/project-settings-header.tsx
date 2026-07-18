import { Link } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';
import { useProjectDetails } from '@/hooks/use-project-details';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/shared/components/button';
import { HeaderActions, HeaderLogo, HeaderShell, HeaderStart } from '@/shared/components/header-shell';

export const ProjectSettingsHeader = ({ projectId }: { projectId: string }) => {
  const { data: project } = useProjectDetails(projectId);

  if (!project) {
    return null;
  }

  return (
    <HeaderShell>
      <HeaderStart>
        <HeaderLogo />
        <div className="flex min-w-0 items-center gap-2 pl-2">
          <Link
            to="/projects/$projectId"
            params={{
              projectId,
            }}
            className={cn(
              buttonVariants({ variant: 'link', size: 'sm' }),
              'text-muted-foreground min-w-0 truncate text-sm',
            )}
          >
            {project.name}
          </Link>
          <span className="text-muted-foreground text-sm">/</span>
          <p className="text-sm font-semibold">Settings</p>
        </div>
      </HeaderStart>

      <HeaderActions>
        <Link
          to="/projects/$projectId"
          params={{ projectId }}
          className={cn(buttonVariants({ variant: 'ghost', size: 'sm' }))}
        >
          <ArrowLeft className="h-4 w-4" />
          Back to board
        </Link>
      </HeaderActions>
    </HeaderShell>
  );
};

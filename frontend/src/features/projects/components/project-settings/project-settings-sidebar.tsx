import { Link } from '@tanstack/react-router';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/shared/components/button';

const sidebarLinkClassName = cn(buttonVariants({ variant: 'ghost' }), 'text-muted-foreground w-full justify-start');
const activeSidebarLinkClassName = cn(sidebarLinkClassName, 'bg-muted text-foreground');

export const ProjectSettingsSidebar = ({ projectId }: { projectId: string }) => {
  return (
    <aside>
      <h1 className="text-xl font-semibold tracking-[-0.01em]">Settings</h1>
      <div className="flex w-full flex-col gap-2 pt-4">
        <Link
          to="/projects/$projectId/settings"
          params={{
            projectId,
          }}
          activeOptions={{
            exact: true,
          }}
          activeProps={{
            className: activeSidebarLinkClassName,
          }}
          inactiveProps={{
            className: sidebarLinkClassName,
          }}
          replace
        >
          General
        </Link>
        <Link
          to="/projects/$projectId/settings/members"
          params={{
            projectId,
          }}
          activeOptions={{
            exact: true,
          }}
          activeProps={{
            className: activeSidebarLinkClassName,
          }}
          inactiveProps={{
            className: sidebarLinkClassName,
          }}
          replace
        >
          Members
        </Link>
        <Link
          to="/projects/$projectId/settings/columns"
          params={{
            projectId,
          }}
          activeOptions={{
            exact: true,
          }}
          activeProps={{
            className: activeSidebarLinkClassName,
          }}
          inactiveProps={{
            className: sidebarLinkClassName,
          }}
          replace
        >
          Columns
        </Link>
      </div>
    </aside>
  );
};

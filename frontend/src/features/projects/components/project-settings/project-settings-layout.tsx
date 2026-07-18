import { Outlet } from '@tanstack/react-router';
import { ProjectSettingsSidebar } from './project-settings-sidebar';
import { ProjectSettingsHeader } from './project-settings-header';

export const ProjectSettingsLayout = ({ projectId }: { projectId: string }) => {
  return (
    <div className="bg-background flex h-screen w-full flex-col overflow-hidden">
      <ProjectSettingsHeader projectId={projectId} />

      <div className="min-h-0 flex-1 overflow-hidden">
        <div className="grid h-full min-h-0 w-full grid-cols-[minmax(24px,1fr)_180px_48px_minmax(0,796px)_minmax(24px,1fr)] pt-12">
          <div className="col-start-2">
            <ProjectSettingsSidebar projectId={projectId} />
          </div>

          <main className="col-start-4 col-end-6 min-h-0 overflow-y-auto">
            <div className="w-[calc(100%-1.5rem)] max-w-[796px]">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </div>
  );
};

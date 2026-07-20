import { ProjectsPageHeader } from './projects-page-header';
import { ProjectsGrid } from './projects-grid';
import { RecentProjectActivity } from './recent-project-activity';
import { UpcomingDeadlines } from './upcoming-deadlines';

export const ProjectsPage = () => {
  return (
    <div className="bg-background min-h-screen font-sans">
      <ProjectsPageHeader />

      <main className="mx-auto w-full max-w-6xl px-4 pt-8 pb-16 sm:px-6 sm:pt-9">
        <ProjectsGrid />

        <div className="mt-8 grid items-stretch gap-5 lg:grid-cols-5">
          <div className="lg:col-span-3">
            <RecentProjectActivity />
          </div>
          <div className="lg:col-span-2">
            <UpcomingDeadlines />
          </div>
        </div>
      </main>
    </div>
  );
};

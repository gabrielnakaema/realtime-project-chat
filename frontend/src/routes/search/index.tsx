import { Link, createFileRoute } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';
import { z } from 'zod';
import { HeaderUser } from '@/components/header-user';
import { NotificationBell } from '@/components/notification-bell';
import { ProjectSearchResults } from '@/components/search/project-search-results';
import { SearchBar } from '@/components/search-bar';
import { TaskSearchResults } from '@/components/search/task-search-results';

export const Route = createFileRoute('/search/')({
  component: RouteComponent,
  validateSearch: z.object({
    query: z.coerce.string().optional(),
  }),
});

function RouteComponent() {
  const { query } = Route.useSearch();

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <div className="px-6 py-4">
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="flex items-center gap-4">
              <Link
                to="/projects"
                className="inline-flex items-center rounded-md px-3 py-2 font-medium whitespace-nowrap text-slate-700 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Go back
              </Link>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">TaskFlow</h1>
            </div>

            <SearchBar action="/search" searchName="query" formClassName="w-full max-w-md" initialValue={query} />
            <div className="flex items-center gap-4">
              <NotificationBell />
              <HeaderUser />
            </div>
          </div>
        </div>
      </header>
      <div className="flex w-full flex-col gap-8 px-6 py-12">
        <p className="text-lg font-medium text-slate-900 dark:text-slate-400">
          Showing results for <span className="text-slate-500 dark:text-slate-100">&quot;{query}&quot;</span>
        </p>
        <ProjectSearchResults query={query} />

        <TaskSearchResults query={query} />
      </div>
    </div>
  );
}

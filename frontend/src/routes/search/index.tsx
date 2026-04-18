import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { HeaderUser } from '@/components/header-user';
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
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">TaskFlow</h1>
            </div>

            <SearchBar action="/search" searchName="query" formClassName="w-full max-w-md" initialValue={query} />
            <div className="flex items-center gap-4">
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

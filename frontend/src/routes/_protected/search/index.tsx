import { Link, createFileRoute } from '@tanstack/react-router';
import { ArrowLeft, Search } from 'lucide-react';
import { z } from 'zod';
import { HeaderUser } from '@/features/auth/components/header-user';
import { NotificationBell } from '@/features/notifications/components/notification-bell';
import { SearchEmptyState } from '@/features/search/components/empty-state';
import { ProjectSearchResults } from '@/features/search/components/project-search-results';
import { SearchBar } from '@/features/search/components/search-bar';
import { TaskSearchResults } from '@/features/search/components/task-search-results';
import { normalizeSearchQuery } from '@/shared/utils/search';

export const Route = createFileRoute('/_protected/search/')({
  component: RouteComponent,
  validateSearch: z.object({
    query: z.coerce.string().optional(),
  }),
});

function RouteComponent() {
  const { query } = Route.useSearch();
  const normalizedQuery = normalizeSearchQuery(query);
  const hasQuery = normalizedQuery.length > 0;

  return (
    <div className="bg-muted min-h-screen">
      <header className="border-border bg-card border-b">
        <div className="px-6 py-4">
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="flex items-center gap-4">
              <Link
                to="/projects"
                className="text-foreground hover:bg-muted inline-flex items-center rounded-md px-3 py-2 font-medium whitespace-nowrap transition-colors"
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Go back
              </Link>
              <h1 className="text-foreground text-2xl font-bold">TaskFlow</h1>
            </div>

            <SearchBar
              key={normalizedQuery}
              action="/search"
              searchName="query"
              formClassName="w-full max-w-md"
              initialValue={normalizedQuery}
            />
            <div className="flex items-center gap-4">
              <NotificationBell />
              <HeaderUser />
            </div>
          </div>
        </div>
      </header>
      <div className="flex w-full flex-col gap-8 px-6 py-12">
        {hasQuery ? (
          <>
            <p className="text-foreground text-lg font-medium">
              Showing results for <span className="text-muted-foreground">&quot;{normalizedQuery}&quot;</span>
            </p>
            <ProjectSearchResults query={normalizedQuery} />
            <TaskSearchResults query={normalizedQuery} />
          </>
        ) : (
          <SearchEmptyState
            icon={<Search className="h-4 w-4" />}
            title="Search across your workspace"
            description="Look up projects and tasks to jump back into ongoing work."
          />
        )}
      </div>
    </div>
  );
}

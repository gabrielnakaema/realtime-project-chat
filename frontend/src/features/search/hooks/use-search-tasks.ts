import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import type { CursorPaginated } from '@/shared/types/paginated';
import type { Task } from '@/features/tasks/types/task';
import { useInfiniteScrollObserver } from '@/shared/hooks/use-infinite-scroll-observer';
import { taskQueryKeys } from '@/shared/services/query-keys';
import { searchTasksForUser } from '@/features/tasks/services/tasks';
import { normalizeSearchQuery } from '@/shared/utils/search';

const PAGE_SIZE = 15;

export const useSearchTasks = (query?: string) => {
  const normalizedQuery = normalizeSearchQuery(query);

  const result = useInfiniteQuery({
    queryKey: taskQueryKeys.search({
      searchQuery: normalizedQuery,
      limit: PAGE_SIZE,
      cursorDueDate: null,
      cursorTaskId: null,
      cursorUpdatedAt: null,
    }),
    queryFn: ({ pageParam }) =>
      searchTasksForUser({
        searchQuery: normalizedQuery,
        limit: PAGE_SIZE,
        cursorDueDate: pageParam.cursorDueDate,
        cursorTaskId: pageParam.cursorTaskId,
        cursorUpdatedAt: pageParam.cursorUpdatedAt,
      }),
    enabled: normalizedQuery.length > 0,
    getNextPageParam: (lastPage: CursorPaginated<Task>) => {
      if (!lastPage.has_next) return undefined;
      const last = lastPage.data[lastPage.data.length - 1];
      return { cursorDueDate: last.due_date, cursorTaskId: last.id, cursorUpdatedAt: last.updated_at };
    },
    initialPageParam: {
      cursorDueDate: null,
      cursorTaskId: null,
      cursorUpdatedAt: '',
    } as {
      cursorDueDate: null | string;
      cursorTaskId: null | string;
      cursorUpdatedAt: string;
    },
  });
  const { fetchNextPage, isFetchingNextPage } = result;

  const data = useMemo(() => result.data?.pages.flatMap((p) => p.data) ?? [], [result.data]);

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
  });

  return {
    data,
    isLoading: result.isLoading,
    isFetchingNextPage,
    sentinelRef,
  };
};

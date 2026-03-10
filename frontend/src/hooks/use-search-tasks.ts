import { useInfiniteQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef } from 'react';
import type { CursorPaginated } from '@/types/paginated';
import type { Task } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';
import { searchTasksForUser } from '@/services/tasks';

const PAGE_SIZE = 15;

export const useSearchTasks = (query?: string) => {
  const sentinelRef = useRef<HTMLDivElement>(null);

  const result = useInfiniteQuery({
    queryKey: taskQueryKeys.search({
      searchQuery: query || '',
      limit: PAGE_SIZE,
      cursorDueDate: null,
      cursorUpdatedAt: null,
    }),
    queryFn: ({ pageParam }) =>
      searchTasksForUser({
        searchQuery: query || '',
        limit: PAGE_SIZE,
        cursorDueDate: pageParam.cursorDueDate,
        cursorUpdatedAt: pageParam.cursorUpdatedAt,
      }),
    getNextPageParam: (lastPage: CursorPaginated<Task>) => {
      if (!lastPage.has_next) return undefined;
      const last = lastPage.data[lastPage.data.length - 1];
      return { cursorDueDate: last.due_date, cursorUpdatedAt: last.updated_at };
    },
    initialPageParam: {
      cursorDueDate: null,
      cursorUpdatedAt: '',
    } as {
      cursorDueDate: null | string;
      cursorUpdatedAt: string;
    },
  });
  const { fetchNextPage, hasNextPage } = result;

  const data = useMemo(() => result.data?.pages.flatMap((p) => p.data) ?? [], [result.data]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || !hasNextPage) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) fetchNextPage();
      },
      { root: null, rootMargin: '200px', threshold: 0.1 },
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, [fetchNextPage, hasNextPage, data.length]);

  return {
    data,
    isLoading: result.isLoading,
    isFetchingNextPage: result.isFetchingNextPage,
    sentinelRef,
  };
};

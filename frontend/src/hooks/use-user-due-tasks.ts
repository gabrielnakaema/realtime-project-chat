import { useInfiniteQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef } from 'react';
import type { CursorPaginated } from '@/types/paginated';
import type { Task } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';
import { listUserDueTasks } from '@/services/tasks';

const PAGE_SIZE = 10;

export const useUserDueTasks = () => {
  const sentinelRef = useRef<HTMLDivElement>(null);

  const query = useInfiniteQuery({
    queryKey: taskQueryKeys.listUserDueTasks,
    initialPageParam: { cursorDueDate: null, cursorUpdatedAt: '' } as {
      cursorDueDate: null | string;
      cursorUpdatedAt: string;
    },
    queryFn: ({ pageParam }) =>
      listUserDueTasks({
        cursorDueDate: pageParam.cursorDueDate,
        cursorUpdatedAt: pageParam.cursorUpdatedAt,
        limit: PAGE_SIZE,
      }),
    getNextPageParam: (lastPage: CursorPaginated<Task>) => {
      if (!lastPage.has_next) return undefined;
      const last = lastPage.data[lastPage.data.length - 1];
      return { cursorDueDate: last.due_date, cursorUpdatedAt: last.updated_at };
    },
  });

  const data = useMemo(() => query.data?.pages.flatMap((p) => p.data) ?? [], [query.data]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || !query.hasNextPage) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) query.fetchNextPage();
      },
      { root: null, rootMargin: '200px', threshold: 0.1 },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [query.fetchNextPage, query.hasNextPage, data.length]);

  return {
    data,
    isLoading: query.isLoading,
    isFetchingNextPage: query.isFetchingNextPage,
    sentinelRef,
  };
};

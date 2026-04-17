import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import type { CursorPaginated } from '@/types/paginated';
import type { Task } from '@/types/task';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { taskQueryKeys } from '@/services/query-keys';
import { listUserDueTasks } from '@/services/tasks';

const PAGE_SIZE = 10;

export const useUserDueTasks = () => {
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

  const { fetchNextPage, isFetchingNextPage } = query;

  const data = useMemo(() => query.data?.pages.flatMap((p) => p.data) ?? [], [query.data]);

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
  });

  return {
    data,
    isLoading: query.isLoading,
    isFetchingNextPage,
    sentinelRef,
  };
};

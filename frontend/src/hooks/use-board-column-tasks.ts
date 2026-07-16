import { useInfiniteQuery } from '@tanstack/react-query';
import { useCallback } from 'react';
import { sortedColumnTasks } from './task-board-cache';
import type { BoardColumnPage, ColumnCache } from './task-board-cache';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { taskQueryKeys } from '@/services/query-keys';
import { listColumnTasks } from '@/services/tasks';

interface UseBoardColumnTasksProps {
  projectId: string;
  columnId: string;
  limit: number;
}

interface PageParam {
  taskOrder: string;
  updatedAt: null | string;
}

export const useBoardColumnTasks = ({ projectId, columnId, limit }: UseBoardColumnTasksProps) => {
  const infiniteQueryResult = useInfiniteQuery({
    // eslint-disable-next-line @tanstack/query/exhaustive-deps
    queryKey: taskQueryKeys.boardColumn(projectId, columnId),
    queryFn: async ({ pageParam }): Promise<BoardColumnPage> => {
      const page = await listColumnTasks({
        projectId,
        columnId,
        archived: false,
        limit,
        taskOrder: pageParam.taskOrder,
        updatedAt: pageParam.updatedAt,
      });

      const lastTask = page.data.at(-1);
      return {
        ...page,
        cursor: lastTask ? { taskOrder: lastTask.order, updatedAt: lastTask.updated_at } : null,
      };
    },
    getNextPageParam: (lastPage): PageParam | undefined =>
      lastPage.has_next && lastPage.cursor ? lastPage.cursor : undefined,
    initialPageParam: {
      taskOrder: '',
      updatedAt: null,
    } as PageParam,
    select: useCallback((data: ColumnCache) => sortedColumnTasks(data), []),
  });
  const { fetchNextPage } = infiniteQueryResult;

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
  });

  return {
    columnTasks: infiniteQueryResult.data ?? [],
    queryResult: infiniteQueryResult,
    sentinelRef,
  };
};

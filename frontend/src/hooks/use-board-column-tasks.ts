import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { taskQueryKeys } from '@/services/query-keys';
import { listGroupedTasksByProjectId } from '@/services/tasks';

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
    queryKey: taskQueryKeys.listGroupedByProjectId({
      projectId,
      projectColumnIds: [columnId],
      archived: false,
      limit,
      taskOrder: '',
      updatedAt: null,
    }),
    queryFn: ({ pageParam }) => {
      return listGroupedTasksByProjectId({
        projectId,
        projectColumnIds: [columnId],
        archived: false,
        taskOrder: pageParam.taskOrder,
        updatedAt: pageParam.updatedAt,
        limit,
      });
    },
    getNextPageParam: (lastPage) => {
      if (!lastPage[columnId].has_next) {
        return undefined;
      }

      const lastTask = lastPage[columnId].data[lastPage[columnId].data.length - 1];

      return {
        taskOrder: lastTask.order,
        updatedAt: lastTask.updated_at,
      };
    },
    initialPageParam: {
      taskOrder: '',
      updatedAt: null,
    } as PageParam,
  });
  const { fetchNextPage } = infiniteQueryResult;

  const columnTasks = useMemo(() => {
    return infiniteQueryResult.data?.pages.flatMap((page) => page[columnId].data) ?? [];
  }, [infiniteQueryResult.data, columnId]);

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
  });

  return {
    columnTasks,
    queryResult: infiniteQueryResult,
    sentinelRef,
  };
};

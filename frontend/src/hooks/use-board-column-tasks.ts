import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import type { TaskStatus } from '@/types/task';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { taskQueryKeys } from '@/services/query-keys';
import { listGroupedTasksByProjectId } from '@/services/tasks';

interface UseBoardColumnTasksProps {
  projectId: string;
  status: TaskStatus;
  limit: number;
}

interface PageParam {
  taskOrder: string;
  updatedAt: null | string;
}

export const useBoardColumnTasks = ({ projectId, status, limit }: UseBoardColumnTasksProps) => {
  const infiniteQueryResult = useInfiniteQuery({
    queryKey: taskQueryKeys.listGroupedByProjectId({
      projectId,
      statuses: [status],
      limit,
      taskOrder: '',
      updatedAt: null,
    }),
    queryFn: ({ pageParam }) => {
      return listGroupedTasksByProjectId({
        projectId,
        statuses: [status],
        taskOrder: pageParam.taskOrder,
        updatedAt: pageParam.updatedAt,
        limit,
      });
    },
    getNextPageParam: (lastPage) => {
      if (!lastPage[status].has_next) {
        return undefined;
      }

      const lastTask = lastPage[status].data[lastPage[status].data.length - 1];

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
    return infiniteQueryResult.data?.pages.flatMap((page) => page[status].data) ?? [];
  }, [infiniteQueryResult.data, status]);

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
  });

  return {
    columnTasks,
    queryResult: infiniteQueryResult,
    sentinelRef,
  };
};

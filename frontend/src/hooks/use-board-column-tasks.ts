import { useInfiniteQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef } from 'react';
import type { TaskStatus } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';
import { listGroupedTasksByProjectId } from '@/services/tasks';

interface UseBoardColumnTasksProps {
  projectId: string;
  status: TaskStatus;
  limit: number;
}

interface PageParam {
  taskOrder: number;
  updatedAt: null | string;
}

export const useBoardColumnTasks = ({ projectId, status, limit }: UseBoardColumnTasksProps) => {
  const sentinelRef = useRef<HTMLDivElement>(null);

  const infiniteQueryResult = useInfiniteQuery({
    queryKey: taskQueryKeys.listGroupedByProjectId({
      projectId,
      statuses: [status],
      limit,
      taskOrder: 0,
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
      taskOrder: 0,
      updatedAt: null,
    } as PageParam,
  });
  const { fetchNextPage, hasNextPage } = infiniteQueryResult;

  const columnTasks = useMemo(() => {
    return infiniteQueryResult.data?.pages.flatMap((page) => page[status].data) ?? [];
  }, [infiniteQueryResult.data, status]);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting && hasNextPage) {
            fetchNextPage();
          }
        });
      },
      {
        root: null,
        rootMargin: '200px',
        threshold: 0.1,
      },
    );

    if (sentinelRef.current) {
      observer.observe(sentinelRef.current);
    }

    return () => {
      observer.disconnect();
    };
  }, [fetchNextPage, hasNextPage]);

  return {
    columnTasks,
    queryResult: infiniteQueryResult,
    sentinelRef,
  };
};

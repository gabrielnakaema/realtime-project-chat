import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { listProjectActivities } from '@/services/project-activities';
import { projectQueryKeys } from '@/services/query-keys';

export const useProjectActivities = (projectId: string) => {
  const infiniteQueryResult = useInfiniteQuery({
    queryKey: projectQueryKeys.activities(projectId),
    initialPageParam: {
      before: new Date().toISOString(),
      id: '',
    },
    queryFn: ({ pageParam }) =>
      listProjectActivities({ projectId, before: pageParam.before, id: pageParam.id, limit: 10 }),
    getNextPageParam: (lastPage) => {
      if (!lastPage.has_next) {
        return undefined;
      }

      const lastPageActivity = lastPage.data[lastPage.data.length - 1];

      return {
        before: lastPageActivity.created_at,
        id: lastPageActivity.id,
      };
    },
  });

  const { fetchNextPage } = infiniteQueryResult;

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: fetchNextPage,
    rootMargin: '100px',
  });

  const data = useMemo(() => {
    return infiniteQueryResult.data?.pages.flatMap((page) => page.data) ?? [];
  }, [infiniteQueryResult.data]);

  return {
    data,
    queryResult: infiniteQueryResult,
    sentinelRef,
  };
};

import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import { listUserProjectActivities } from '@/features/projects/services/project-activities';
import { useInfiniteScrollObserver } from '@/shared/hooks/use-infinite-scroll-observer';

export const useUserProjectActivities = () => {
  const infiniteQueryResult = useInfiniteQuery({
    queryKey: ['project-activities', 'user'],
    initialPageParam: {
      before: new Date().toISOString(),
      id: '',
    },
    queryFn: ({ pageParam }) => listUserProjectActivities({ before: pageParam.before, id: pageParam.id, limit: 10 }),
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

import { useInfiniteQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef } from 'react';
import { listProjectActivities } from '@/services/project-activities';
import { projectQueryKeys } from '@/services/query-keys';

export const useProjectActivities = (projectId: string) => {
  const sentinelRef = useRef<HTMLDivElement>(null);

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

  const { fetchNextPage, hasNextPage } = infiniteQueryResult;

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
        rootMargin: '100px',
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

  const data = useMemo(() => {
    return infiniteQueryResult.data?.pages.flatMap((page) => page.data) ?? [];
  }, [infiniteQueryResult.data]);

  return {
    data,
    queryResult: infiniteQueryResult,
    sentinelRef,
  };
};

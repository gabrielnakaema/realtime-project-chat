import { useInfiniteQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import { listUserProjectActivities } from '@/services/project-activities';

export const useUserProjectActivities = () => {
  const infiniteQueryResult = useInfiniteQuery({
    queryKey: ['project-activities', 'user'],
    initialPageParam: {
      before: '',
      id: '',
    },
    queryFn: ({ pageParam }) => listUserProjectActivities({ before: pageParam.before, id: pageParam.id, limit: 10 }),
    getNextPageParam: (lastPage) => {
      if (!lastPage.has_next) {
        return undefined;
      }

      return {
        before: lastPage.data[0].created_at,
        id: lastPage.data[0].id,
      };
    },
  });

  const data = useMemo(() => {
    return infiniteQueryResult.data?.pages.flatMap((page) => page.data) ?? [];
  }, [infiniteQueryResult.data]);

  return {
    data,
    queryResult: infiniteQueryResult,
  };
};

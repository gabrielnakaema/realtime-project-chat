import type { QueryClient } from '@tanstack/react-query';
import { projectQueryKeys, taskQueryKeys } from '@/shared/services/query-keys';

export const invalidateProjectBoardData = async (queryClient: QueryClient) => {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: projectQueryKeys.all }),
    queryClient.invalidateQueries({ queryKey: taskQueryKeys.all }),
  ]);
};

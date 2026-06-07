import { projectQueryKeys, taskQueryKeys } from './query-keys';
import type { QueryClient } from '@tanstack/react-query';

export const invalidateProjectBoardData = async (queryClient: QueryClient) => {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: projectQueryKeys.all }),
    queryClient.invalidateQueries({ queryKey: taskQueryKeys.all }),
  ]);
};

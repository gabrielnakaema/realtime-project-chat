import { useQuery } from '@tanstack/react-query';
import { listProjects } from '@/features/projects/services/projects';
import { projectQueryKeys } from '@/shared/services/query-keys';

export const useListProjects = () => {
  return useQuery({
    queryKey: projectQueryKeys.list,
    queryFn: () => listProjects(),
  });
};

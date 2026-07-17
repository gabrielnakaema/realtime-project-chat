import { useQuery } from '@tanstack/react-query';
import { listProjects } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';

export const useListProjects = () => {
  return useQuery({
    queryKey: projectQueryKeys.list,
    queryFn: () => listProjects(),
  });
};

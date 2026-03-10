import { useQuery } from '@tanstack/react-query';
import { listProjects } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';

export const useSearchProjects = (query?: string) => {
  const { data, isLoading } = useQuery({
    queryKey: projectQueryKeys.search(query),
    queryFn: () => listProjects(query),
  });

  return {
    data,
    isLoading,
  };
};

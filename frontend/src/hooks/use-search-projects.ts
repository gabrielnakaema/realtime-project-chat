import { useQuery } from '@tanstack/react-query';
import { listProjects } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';
import { normalizeSearchQuery } from '@/utils/search';

export const useSearchProjects = (query?: string) => {
  const normalizedQuery = normalizeSearchQuery(query);

  const { data, isLoading } = useQuery({
    queryKey: projectQueryKeys.search(normalizedQuery),
    queryFn: () => listProjects(normalizedQuery),
    enabled: normalizedQuery.length > 0,
  });

  return {
    data,
    isLoading,
  };
};

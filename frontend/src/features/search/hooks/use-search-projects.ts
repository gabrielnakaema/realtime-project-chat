import { useQuery } from '@tanstack/react-query';
import { listProjects } from '@/features/projects/services/projects';
import { projectQueryKeys } from '@/shared/services/query-keys';
import { normalizeSearchQuery } from '@/shared/utils/search';

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

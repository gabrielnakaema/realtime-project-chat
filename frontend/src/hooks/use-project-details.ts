import { useQuery } from '@tanstack/react-query';
import type { UseQueryOptions } from '@tanstack/react-query';
import type { Project } from '@/types/project';
import { getProject } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';

export const useProjectDetails = (projectId: string, options?: Partial<UseQueryOptions<Project>>) => {
  return useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
    ...options,
  });
};

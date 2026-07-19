import { useQuery } from '@tanstack/react-query';
import type { UseQueryOptions } from '@tanstack/react-query';
import type { Project } from '@/features/projects/types/project';
import { getProject } from '@/features/projects/services/projects';
import { projectQueryKeys } from '@/shared/services/query-keys';

export const useProjectDetails = (projectId: string, options?: Partial<UseQueryOptions<Project>>) => {
  return useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
    ...options,
  });
};

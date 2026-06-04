import { useQuery } from '@tanstack/react-query';
import { listMembersByProjectId } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';

const PROJECT_MEMBERS_STALE_TIME = 10 * 1000; // 10 seconds

export const useProjectMembers = (projectId: string, enabled = true) => {
  return useQuery({
    queryKey: projectQueryKeys.members(projectId),
    queryFn: () => listMembersByProjectId(projectId),
    enabled: enabled && !!projectId,
    staleTime: PROJECT_MEMBERS_STALE_TIME,
  });
};

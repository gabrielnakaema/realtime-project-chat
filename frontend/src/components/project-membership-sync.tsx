import { useQueryClient } from '@tanstack/react-query';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useEffect, useEffectEvent } from 'react';
import type { SocketEvent } from '@/types/websocket';
import { isProjectMembershipEvent } from '@/types/websocket';
import { useAuth } from '@/hooks/use-auth';
import { useSocket } from '@/hooks/use-socket';
import { invalidateProjectBoardData } from '@/services/project-board-invalidation';
import { generalChatQueryKeys, projectChatQueryKeys, projectQueryKeys, taskQueryKeys } from '@/services/query-keys';

export const ProjectMembershipSync = () => {
  const { user } = useAuth();
  const { status, subscribe } = useSocket();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });

  const handleMembershipEvent = useEffectEvent((event: SocketEvent) => {
    if (!isProjectMembershipEvent(event)) {
      return;
    }

    void invalidateProjectBoardData(queryClient);
    queryClient.invalidateQueries({ queryKey: projectChatQueryKeys.all });
    queryClient.invalidateQueries({ queryKey: generalChatQueryKeys.list });

    if (event.type === 'project_member_removed') {
      // Purge on top of the invalidation above so the removed project/tasks
      // disappear immediately instead of refetching into a 403.
      queryClient.removeQueries({ queryKey: projectQueryKeys.details(event.data.project_id) });
      queryClient.removeQueries({ queryKey: taskQueryKeys.all });

      if (pathname.startsWith(`/projects/${event.data.project_id}`)) {
        void navigate({ to: '/projects' });
      }
    }
  });

  useEffect(() => {
    if (!user?.id || status !== 'connected') {
      return;
    }

    return subscribe(user.id, 'user', handleMembershipEvent);
  }, [status, subscribe, user?.id]);

  return null;
};

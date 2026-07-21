import { useQueryClient } from '@tanstack/react-query';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useEffect, useEffectEvent } from 'react';
import { toast } from 'react-toastify';
import type { SocketEvent } from '@/shared/types/websocket';
import { isProjectDeletedEvent, isProjectMembershipEvent } from '@/shared/types/websocket';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useSocket } from '@/shared/hooks/use-socket';
import { invalidateProjectBoardData } from '@/features/projects/services/project-board-invalidation';
import {
  generalChatQueryKeys,
  projectChatQueryKeys,
  projectQueryKeys,
  taskQueryKeys,
} from '@/shared/services/query-keys';

const PROJECT_ROUTE_PATTERN = /^\/projects\/([^/]+)/;

export const ProjectSync = () => {
  const { user } = useAuth();
  const { status, subscribe } = useSocket();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const currentProjectId = pathname.match(PROJECT_ROUTE_PATTERN)?.[1];

  const handleMembershipEvent = useEffectEvent((event: SocketEvent) => {
    if (!isProjectMembershipEvent(event)) {
      return;
    }

    void invalidateProjectBoardData(queryClient);
    queryClient.invalidateQueries({ queryKey: projectChatQueryKeys.all });
    queryClient.invalidateQueries({ queryKey: generalChatQueryKeys.list });

    if (event.type === 'project_member_removed') {
      queryClient.removeQueries({ queryKey: projectQueryKeys.details(event.data.project_id) });
      queryClient.removeQueries({ queryKey: taskQueryKeys.all });

      if (pathname.startsWith(`/projects/${event.data.project_id}`)) {
        void navigate({ to: '/projects' });
      }
    }
  });

  const handleProjectDeletedEvent = useEffectEvent((event: SocketEvent) => {
    if (!isProjectDeletedEvent(event)) {
      return;
    }

    void invalidateProjectBoardData(queryClient);
    queryClient.invalidateQueries({ queryKey: projectChatQueryKeys.all });
    queryClient.invalidateQueries({ queryKey: generalChatQueryKeys.list });
    queryClient.removeQueries({ queryKey: projectQueryKeys.details(event.data.id) });
    queryClient.removeQueries({ queryKey: taskQueryKeys.all });

    toast.info('This project was deleted.');
    void navigate({ to: '/projects' });
  });

  useEffect(() => {
    if (!user?.id || status !== 'connected') {
      return;
    }

    return subscribe(user.id, 'user', handleMembershipEvent);
  }, [status, subscribe, user?.id]);

  useEffect(() => {
    if (!currentProjectId || status !== 'connected') {
      return;
    }

    return subscribe(currentProjectId, 'project', handleProjectDeletedEvent);
  }, [currentProjectId, status, subscribe]);

  return null;
};

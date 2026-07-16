import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useRef } from 'react';
import { useSocket } from './use-socket';
import { invalidateBoard, reconcileTask } from './task-board-cache';
import type { SocketEvent } from '@/types/websocket';
import { isProjectMembershipEvent } from '@/types/websocket';
import { projectQueryKeys } from '@/services/query-keys';

export const useRealtimeTaskSync = (projectId: string) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();
  const connectedOnce = useRef(false);

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type === 'project_updated') {
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.details(projectId) });
      return;
    }

    if (isProjectMembershipEvent(event)) {
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.details(projectId) });
      return;
    }

    if (event.type === 'task_created') {
      reconcileTask(queryClient, event.data);
      return;
    }

    if (event.type === 'task_updated') {
      const { task, previous_project_column_id } = event.data;
      reconcileTask(queryClient, task, { previousColumnId: previous_project_column_id });
    }
  });

  useEffect(() => {
    if (!projectId || status !== 'connected') return;
    const unsubscribe = subscribe(projectId, 'project', handleSocketEvent);
    return () => unsubscribe();
  }, [projectId, status, subscribe]);

  useEffect(() => {
    if (status !== 'connected') return;
    if (connectedOnce.current) {
      invalidateBoard(queryClient, projectId);
    }
    connectedOnce.current = true;
  }, [status, projectId, queryClient]);
};

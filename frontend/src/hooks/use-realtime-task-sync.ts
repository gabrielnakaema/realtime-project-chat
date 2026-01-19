import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent } from 'react';
import { useSocket } from './use-socket';
import type { SocketEvent } from '@/types/websocket';
import { taskQueryKeys } from '@/services/query-keys';

export const useRealtimeTaskSync = (projectId: string) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type === 'task_updated' || event.type === 'task_created') {
      queryClient.invalidateQueries({
        queryKey: taskQueryKeys.all,
      });
    }
  });

  useEffect(() => {
    if (!projectId || status !== 'connected') {
      return;
    }

    const unsubscribe = subscribe(projectId, 'project', handleSocketEvent);

    return () => unsubscribe();
  }, [projectId, status, subscribe]);
};

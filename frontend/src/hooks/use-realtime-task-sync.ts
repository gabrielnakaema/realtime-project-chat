import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useMemo, useRef } from 'react';
import { useSocket } from './use-socket';
import { buildColumnQueryKey, reconcileTaskInBoard } from './task-cache-helpers';
import type { SocketEvent } from '@/types/websocket';
import { isProjectMembershipEvent } from '@/types/websocket';
import { projectQueryKeys, taskQueryKeys } from '@/services/query-keys';

export const useRealtimeTaskSync = (projectId: string, projectColumnIds: string[]) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();
  const connectedOnce = useRef(false);
  const columnIdsKey = projectColumnIds.join(',');
  const stableProjectColumnIds = useMemo(() => {
    return columnIdsKey ? columnIdsKey.split(',') : [];
  }, [columnIdsKey]);

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
      reconcileTaskInBoard(queryClient, projectId, event.data);
      return;
    }

    if (event.type === 'task_updated') {
      const { task, previous_project_column_id } = event.data;
      reconcileTaskInBoard(queryClient, projectId, task, { previousColumnId: previous_project_column_id });
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
      for (const s of stableProjectColumnIds) {
        queryClient.invalidateQueries({ queryKey: buildColumnQueryKey(projectId, s) });
      }
      queryClient.invalidateQueries({
        queryKey: taskQueryKeys.countByColumn(projectId, stableProjectColumnIds),
      });
    }
    connectedOnce.current = true;
  }, [status, projectId, queryClient, stableProjectColumnIds]);
};

import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useRef } from 'react';
import { useSocket } from './use-socket';
import {
  adjustCountCache,
  buildColumnQueryKey,
  insertTaskAtCorrectPosition,
  removeTaskFromColumn,
  updateTaskInColumn,
} from './task-cache-helpers';
import type { SocketEvent } from '@/types/websocket';
import { taskQueryKeys } from '@/services/query-keys';

export const useRealtimeTaskSync = (projectId: string, projectColumnIds: string[]) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();
  const connectedOnce = useRef(false);

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type === 'task_created') {
      const task = event.data;
      insertTaskAtCorrectPosition(queryClient, projectId, task.project_column_id, task);
      adjustCountCache(queryClient, projectId, projectColumnIds, { [task.project_column_id]: 1 });
      return;
    }

    if (event.type === 'task_updated') {
      const { task, previous_project_column_id } = event.data;
      const nextColumn = task.project_column_id;

      if (!previous_project_column_id) {
        updateTaskInColumn(queryClient, projectId, nextColumn, task);
        return;
      }

      const prevColumn = previous_project_column_id;
      removeTaskFromColumn(queryClient, projectId, prevColumn, task.id);

      if (task.archived_at) {
        adjustCountCache(queryClient, projectId, projectColumnIds, { [prevColumn]: -1 });
      } else {
        insertTaskAtCorrectPosition(queryClient, projectId, nextColumn, task);
        adjustCountCache(queryClient, projectId, projectColumnIds, { [prevColumn]: -1, [nextColumn]: 1 });
      }
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
      for (const s of projectColumnIds) {
        queryClient.invalidateQueries({ queryKey: buildColumnQueryKey(projectId, s) });
      }
      queryClient.invalidateQueries({
        queryKey: taskQueryKeys.countByColumn(projectId, projectColumnIds),
      });
    }
    connectedOnce.current = true;
  }, [status, projectId, projectColumnIds, queryClient]);
};

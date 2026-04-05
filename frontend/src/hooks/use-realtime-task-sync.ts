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
import type { TaskStatus } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';
import { TASK_STATUSES } from '@/constants/tasks';

export const useRealtimeTaskSync = (projectId: string) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();
  const connectedOnce = useRef(false);

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type === 'task_created') {
      const task = event.data;
      insertTaskAtCorrectPosition(queryClient, projectId, 'pending', task);
      adjustCountCache(queryClient, projectId, { pending: 1 });
      return;
    }

    if (event.type === 'task_updated') {
      const { task, previous_status } = event.data;
      const nextStatus = task.status;

      if (!previous_status) {
        updateTaskInColumn(queryClient, projectId, nextStatus, task);
        return;
      }

      const prevStatus = previous_status as TaskStatus;
      removeTaskFromColumn(queryClient, projectId, prevStatus, task.id);

      if (nextStatus === 'archived') {
        adjustCountCache(queryClient, projectId, { [prevStatus]: -1 });
      } else {
        insertTaskAtCorrectPosition(queryClient, projectId, nextStatus, task);
        adjustCountCache(queryClient, projectId, { [prevStatus]: -1, [nextStatus]: 1 });
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
      for (const s of TASK_STATUSES) {
        queryClient.invalidateQueries({ queryKey: buildColumnQueryKey(projectId, s) });
      }
      queryClient.invalidateQueries({
        queryKey: taskQueryKeys.countByStatus(projectId, [...TASK_STATUSES]),
      });
    }
    connectedOnce.current = true;
  }, [status, projectId, queryClient]);
};

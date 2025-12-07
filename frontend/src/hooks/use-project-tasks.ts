import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent } from 'react';
import { useSocket } from './use-socket';
import type { Paginated } from '@/types/paginated';
import type { Task } from '@/types/task';
import type { SocketEvent } from '@/types/websocket';
import { countTasksByStatus, listGroupedTasksByProjectId } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';

export const useProjectTasks = (projectId: string) => {
  const queryClient = useQueryClient();

  const params = {
    projectId,
    statuses: ['pending', 'doing', 'done'],
    taskOrder: 0,
    limit: 50,
  };

  const { data, isLoading } = useQuery({
    queryKey: taskQueryKeys.listGroupedByProjectId(params),
    queryFn: () => listGroupedTasksByProjectId(params),
  });

  const { data: countData, isLoading: isCountLoading } = useQuery({
    queryKey: taskQueryKeys.countByStatus(projectId, ['pending', 'doing', 'done']),
    queryFn: () => countTasksByStatus(projectId, ['pending', 'doing', 'done']),
  });

  const { status, subscribe } = useSocket();

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type === 'task_updated') {
      const task = event.data;

      queryClient.setQueryData(taskQueryKeys.listByProjectId(projectId), (old: Paginated<Task>) => {
        const updated = old.data.map((t) => (t.id === task.id ? task : t));

        return {
          ...old,
          data: updated,
        };
      });
    }

    if (event.type === 'task_created') {
      const task = event.data;

      queryClient.setQueryData(taskQueryKeys.listByProjectId(projectId), (old: Paginated<Task>) => {
        return {
          ...old,
          data: [...old.data, task],
        };
      });
    }
  });

  useEffect(() => {
    if (!projectId || status !== 'connected') {
      return;
    }

    const unsubscribe = subscribe(projectId, 'project', handleSocketEvent);

    return () => {
      unsubscribe();
    };
  }, [projectId, status, subscribe]);

  return {
    data,
    isLoading,
    countData,
    isCountLoading,
  };
};

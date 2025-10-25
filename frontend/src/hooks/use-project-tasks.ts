import { useEffect, useEffectEvent } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useSocket } from './use-socket';
import type { Task } from '@/types/task';
import type { SocketEvent } from '@/types/websocket';
import type { Paginated } from '@/types/paginated';
import { taskQueryKeys } from '@/services/query-keys';
import { listTasksByProjectId } from '@/services/tasks';

export const useProjectTasks = (projectId: string) => {
  const queryClient = useQueryClient();

  const { data } = useQuery({
    queryKey: taskQueryKeys.listByProjectId(projectId),
    queryFn: () => listTasksByProjectId(projectId),
  });

  const tasks = data?.data || [];

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
    tasks,
  };
};

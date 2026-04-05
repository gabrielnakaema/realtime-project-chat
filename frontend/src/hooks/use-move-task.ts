import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  findTaskInColumnCaches,
  insertTaskAtCorrectPosition,
  removeTaskFromColumn,
} from './task-cache-helpers';
import { taskQueryKeys } from '@/services/query-keys';
import { moveTask } from '@/services/tasks';

export const useMoveTask = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: moveTask,
    onMutate: async (variables) => {
      const found = findTaskInColumnCaches(queryClient, variables.projectId, variables.taskId);
      if (!found || found.status === variables.status) return;

      const { task: taskSnapshot, status: sourceStatus } = found;

      await queryClient.cancelQueries({ queryKey: taskQueryKeys._allGrouped() });

      removeTaskFromColumn(queryClient, variables.projectId, sourceStatus, variables.taskId);
      insertTaskAtCorrectPosition(queryClient, variables.projectId, variables.status, {
        ...taskSnapshot,
        status: variables.status,
      });

      return { taskSnapshot, sourceStatus };
    },
    onError: (_err, variables, context) => {
      if (!context) return;
      const { taskSnapshot, sourceStatus } = context;
      removeTaskFromColumn(queryClient, variables.projectId, variables.status, variables.taskId);
      insertTaskAtCorrectPosition(queryClient, variables.projectId, sourceStatus, taskSnapshot);
    },
    onSuccess: (updatedTask, variables) => {
      insertTaskAtCorrectPosition(queryClient, variables.projectId, variables.status, updatedTask);
    },
  });
};

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { findTaskInBoardCaches, reconcileTaskInBoard } from './task-cache-helpers';
import { taskQueryKeys } from '@/services/query-keys';
import { moveTask } from '@/services/tasks';

export const useMoveTask = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: moveTask,
    onMutate: async (variables) => {
      const found = findTaskInBoardCaches(queryClient, variables.projectId, variables.taskId);
      if (!found) return;

      const { task: taskSnapshot, columnId: sourceColumnId } = found;

      await queryClient.cancelQueries({ queryKey: taskQueryKeys._allGrouped() });

      reconcileTaskInBoard(
        queryClient,
        variables.projectId,
        { ...taskSnapshot, project_column_id: variables.projectColumnId },
        { force: true },
      );

      return { taskSnapshot, sourceColumnId };
    },
    onError: (_err, variables, context) => {
      if (!context) return;
      reconcileTaskInBoard(queryClient, variables.projectId, context.taskSnapshot, { force: true });
    },
    onSuccess: (updatedTask, variables) => {
      reconcileTaskInBoard(queryClient, variables.projectId, updatedTask);
    },
  });
};

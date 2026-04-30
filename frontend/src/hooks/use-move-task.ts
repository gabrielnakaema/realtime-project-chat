import { useMutation, useQueryClient } from '@tanstack/react-query';
import { findTaskInColumnCaches, insertTaskAtCorrectPosition, removeTaskFromColumn } from './task-cache-helpers';
import { taskQueryKeys } from '@/services/query-keys';
import { moveTask } from '@/services/tasks';

export const useMoveTask = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: moveTask,
    onMutate: async (variables) => {
      const projectColumnIds = variables.projectColumnIds;
      const found = findTaskInColumnCaches(queryClient, variables.projectId, projectColumnIds, variables.taskId);
      if (!found || found.columnId === variables.projectColumnId) return;

      const { task: taskSnapshot, columnId: sourceColumnId } = found;

      await queryClient.cancelQueries({ queryKey: taskQueryKeys._allGrouped() });

      removeTaskFromColumn(queryClient, variables.projectId, sourceColumnId, variables.taskId);
      insertTaskAtCorrectPosition(queryClient, variables.projectId, variables.projectColumnId, {
        ...taskSnapshot,
        project_column_id: variables.projectColumnId,
      });

      return { taskSnapshot, sourceColumnId };
    },
    onError: (_err, variables, context) => {
      if (!context) return;
      const { taskSnapshot, sourceColumnId } = context;
      removeTaskFromColumn(queryClient, variables.projectId, variables.projectColumnId, variables.taskId);
      insertTaskAtCorrectPosition(queryClient, variables.projectId, sourceColumnId, taskSnapshot);
    },
    onSuccess: (updatedTask, variables) => {
      insertTaskAtCorrectPosition(queryClient, variables.projectId, variables.projectColumnId, updatedTask);
    },
  });
};

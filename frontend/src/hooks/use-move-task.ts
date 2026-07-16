import { useMutation, useQueryClient } from '@tanstack/react-query';
import { findTaskOnBoard, invalidateBoard, reconcileTask, sortedColumnTasks } from './task-board-cache';
import type { ColumnCache } from './task-board-cache';
import type { QueryClient } from '@tanstack/react-query';
import { generateKeyBetween } from '@/lib/fracindex';
import { taskQueryKeys } from '@/services/query-keys';
import { moveTask } from '@/services/tasks';

const computeProvisionalOrder = (
  queryClient: QueryClient,
  projectId: string,
  columnId: string,
  taskId: string,
  afterTaskId: string | null,
): string | undefined => {
  const cache = queryClient.getQueryData<ColumnCache>(taskQueryKeys.boardColumn(projectId, columnId));
  const neighbors = sortedColumnTasks(cache).filter((t) => t.id !== taskId);

  let left = '';
  let right = '';
  if (afterTaskId === null) {
    right = neighbors[0]?.order ?? '';
  } else {
    const afterIndex = neighbors.findIndex((t) => t.id === afterTaskId);
    if (afterIndex === -1) {
      left = neighbors.at(-1)?.order ?? '';
    } else {
      left = neighbors[afterIndex].order;
      right = neighbors[afterIndex + 1]?.order ?? '';
    }
  }

  try {
    return generateKeyBetween(left, right);
  } catch {
    return undefined;
  }
};

export const useMoveTask = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: moveTask,
    onMutate: async (variables) => {
      const found = findTaskOnBoard(queryClient, variables.projectId, variables.taskId);
      if (!found) return { sourceColumnId: undefined };

      const { task: taskSnapshot, columnId: sourceColumnId } = found;

      await queryClient.cancelQueries({ queryKey: taskQueryKeys.board(variables.projectId) });

      const provisionalOrder = computeProvisionalOrder(
        queryClient,
        variables.projectId,
        variables.projectColumnId,
        variables.taskId,
        variables.afterTaskId,
      );

      reconcileTask(
        queryClient,
        {
          ...taskSnapshot,
          project_column_id: variables.projectColumnId,
          order: provisionalOrder ?? taskSnapshot.order,
        },
        { optimistic: true },
      );

      return { sourceColumnId };
    },
    onError: (_err, variables) => {
      invalidateBoard(queryClient, variables.projectId);
    },
    onSuccess: (updatedTask, _variables, context) => {
      reconcileTask(queryClient, updatedTask);
      if (context.sourceColumnId !== undefined && context.sourceColumnId !== updatedTask.project_column_id) {
        queryClient.invalidateQueries({ queryKey: taskQueryKeys.counts(updatedTask.project_id) });
      }
    },
  });
};

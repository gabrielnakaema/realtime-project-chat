import { useMutation, useQueryClient } from '@tanstack/react-query';
import { moveTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';

export const useMoveTask = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: moveTask,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.all });
    },
  });
};

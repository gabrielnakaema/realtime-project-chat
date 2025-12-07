import { useMutation, useQueryClient } from '@tanstack/react-query';
import { moveTask } from '@/services/tasks';

export const useMoveTask = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: moveTask,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
};

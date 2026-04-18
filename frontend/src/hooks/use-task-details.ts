import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { archiveTask, getTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';

export const useTaskDetails = (taskId: string, open: boolean, onClose: () => void) => {
  const queryClient = useQueryClient();

  const { data: task, isLoading } = useQuery({
    queryKey: taskQueryKeys.details(taskId),
    queryFn: () => getTask(taskId),
    enabled: open,
  });

  const { mutate: archive, isPending: isArchiving } = useMutation({
    mutationFn: () => archiveTask(taskId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allGrouped() });
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allCounts() });
      onClose();
    },
  });

  return {
    task,
    isLoading,
    archive,
    isArchiving,
  };
};

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { reconcileTask } from './task-board-cache';
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
    onSuccess: (archivedTask) => {
      reconcileTask(queryClient, archivedTask);
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.archived(archivedTask.project_id) });
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

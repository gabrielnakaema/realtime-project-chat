import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { format, parseISO } from 'date-fns';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '../ui/dialog';
import { TaskFormFields } from './task-form-fields';
import { parseUniqueTags } from './task-form-utils';
import type { SubmitHandler } from 'react-hook-form';
import type { ITaskForm } from '@/schemas/task-schema';
import { handleSuccess } from '@/utils/handle-success';
import { getTask, updateTask } from '@/services/tasks';
import { projectQueryKeys, taskQueryKeys } from '@/services/query-keys';
import { listMembersByProjectId } from '@/services/projects';
import { taskSchema } from '@/schemas/task-schema';

interface EditTaskProps {
  taskId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function useEditTaskForm(taskId: string, open: boolean, onOpenChange: (open: boolean) => void) {
  const queryClient = useQueryClient();

  const { data: task, isLoading } = useQuery({
    queryKey: taskQueryKeys.details(taskId),
    queryFn: () => getTask(taskId),
    enabled: open,
  });

  const { data: projectMembers, isLoading: isProjectMembersLoading } = useQuery({
    queryKey: projectQueryKeys.members(task?.project_id ?? ''),
    queryFn: () => listMembersByProjectId(task?.project_id ?? ''),
    enabled: open && !!task?.project_id,
  });

  const memberOptions =
    projectMembers?.map((member) => ({
      label: member.user.name,
      value: member.user.id,
    })) ?? [];

  const {
    register,
    control,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<ITaskForm>({
    resolver: zodResolver(taskSchema),
  });

  const { mutate: submitTask, isPending } = useMutation({
    mutationFn: updateTask,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.all });
      handleSuccess('Task updated successfully');
      onOpenChange(false);
      reset();
    },
  });

  useEffect(() => {
    if (task) {
      reset({
        title: task.title,
        description: task.description,
        due_date: task.due_date ? format(parseISO(task.due_date), 'yyyy-MM-dd') : undefined,
        priority: task.priority,
        responsible_id: task.responsible_id,
        tags: task.tags?.join(',') || undefined,
      });
    }
  }, [task, reset]);

  const onSubmit: SubmitHandler<ITaskForm> = (data) => {
    submitTask({
      id: taskId,
      description: data.description,
      title: data.title,
      priority: data.priority,
      due_date: data.due_date ?? null,
      responsible_id: data.responsible_id ?? null,
      tags: parseUniqueTags(data.tags),
      done_at: null,
      status: task?.status ?? 'pending',
    });
  };

  return {
    task,
    isLoading,
    isProjectMembersLoading,
    memberOptions,
    register,
    control,
    handleSubmit,
    setValue,
    errors,
    onSubmit,
    isPending,
  };
}

export const EditTask = ({ taskId, open, onOpenChange }: EditTaskProps) => {
  const {
    task,
    isLoading,
    isProjectMembersLoading,
    memberOptions,
    register,
    control,
    handleSubmit,
    setValue,
    errors,
    onSubmit,
    isPending,
  } = useEditTaskForm(taskId, open, onOpenChange);

  if (isLoading || isProjectMembersLoading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <div className="flex min-h-50 flex-col items-center justify-center">
            <LoadingSpinner size="3rem" />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  if (!task) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <div className="flex min-h-50 flex-col items-center justify-center">
            <p className="text-sm text-slate-500 dark:text-slate-400">Task not found</p>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="md:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit task</DialogTitle>
          <DialogDescription>Edit the task details</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="flex w-full flex-col gap-4">
          <TaskFormFields
            control={control}
            register={register}
            setValue={setValue}
            errors={errors}
            memberOptions={memberOptions}
            descriptionInitialValue={task.description}
            descriptionEditorKey={task.id}
          />
          <div className="flex w-full items-center justify-end gap-4">
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                Cancel
              </Button>
            </DialogClose>
            <Button type="submit" disabled={isPending}>
              {isPending ? <LoadingSpinner size="1.5em" /> : 'Save changes'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};

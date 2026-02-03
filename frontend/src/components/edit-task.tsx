import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { format, parseISO } from 'date-fns';
import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Button } from './button';
import { Input } from './input';
import { LoadingSpinner } from './loading';
import { Select } from './select';
import { Textarea } from './textarea';
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle } from './ui/dialog';
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

const priorityOptions = [
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
];

export const EditTask = ({ taskId, open, onOpenChange }: EditTaskProps) => {
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

  const memberOptions = projectMembers?.map((member) => ({
    label: member.user.name,
    value: member.user.id,
  }));

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ITaskForm>({
    resolver: zodResolver(taskSchema),
  });

  const { mutate, isPending } = useMutation({
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
    mutate({
      id: taskId,
      description: data.description,
      title: data.title,
      priority: data.priority,
      due_date: data.due_date ?? null,
      responsible_id: data.responsible_id ?? null,
      tags: data.tags?.split(',').map((tag) => tag.trim()) || [],
      done_at: null,
      status: task?.status ?? 'pending',
    });
  };

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
          <DialogTitle>Create task</DialogTitle>
          <DialogDescription>Create a new task for the project</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="flex w-full flex-col gap-4">
          <Input
            label="Title"
            id="title"
            placeholder="Enter task title"
            {...register('title')}
            error={errors.title?.message}
          />
          <Controller
            control={control}
            name="responsible_id"
            render={({ field }) => (
              <Select
                options={memberOptions ?? []}
                value={field.value ?? ''}
                onChange={field.onChange}
                label="Responsible"
                error={errors.responsible_id?.message}
                id="responsible_id"
                placeholder="Select responsible"
              />
            )}
          />
          <div className="flex w-full gap-4">
            <Controller
              control={control}
              name="priority"
              render={({ field }) => (
                <Select
                  options={priorityOptions}
                  value={field.value}
                  onChange={field.onChange}
                  label="Priority"
                  error={errors.priority?.message}
                  id="priority"
                  placeholder="Select priority"
                />
              )}
            />
            <Input
              label="Due date"
              id="due_date"
              type="date"
              placeholder="Enter due date"
              {...register('due_date')}
              error={errors.due_date?.message}
            />
          </div>
          <Controller
            control={control}
            name="tags"
            render={({ field }) => {
              const tags = field.value ?? '';
              const splitTags = tags
                .split(',')
                .map((tag) => tag.trim())
                .filter((tag) => tag !== '');
              const uniqueTags = Array.from(new Set(splitTags));

              return (
                <>
                  <Input
                    label="Tags"
                    id="tags"
                    placeholder="Enter comma separated tags"
                    value={field.value ?? ''}
                    onChange={field.onChange}
                    ref={field.ref}
                    error={errors.tags?.message}
                  />
                  <div className="flex w-full flex-wrap gap-2">
                    {uniqueTags.map((tag) => (
                      <div
                        key={tag}
                        className="w-fit rounded-sm border border-slate-200 px-2 py-0.5 text-xs font-medium text-slate-500 dark:border-slate-700 dark:text-slate-400"
                      >
                        {tag}
                      </div>
                    ))}
                  </div>
                </>
              );
            }}
          />
          <Textarea
            label="Description"
            id="description"
            placeholder="Enter task description"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleSubmit(onSubmit)();
              }
            }}
            {...register('description')}
            error={errors.description?.message}
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

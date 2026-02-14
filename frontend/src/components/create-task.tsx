import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Button } from './button';
import { Input } from './input';
import { LoadingSpinner } from './loading';
import { Select } from './select';
import { TextEditor } from './text-editor';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog';
import type { SubmitHandler } from 'react-hook-form';
import type { Member } from '@/types/project';
import type { ITaskForm } from '@/schemas/task-schema';
import { handleSuccess } from '@/utils/handle-success';
import { createTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { taskSchema } from '@/schemas/task-schema';

interface CreateTaskModalProps {
  projectId: string;
  projectMembers: Member[];
}

const priorityOptions = [
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
];

export const CreateTask = ({ projectId, projectMembers }: CreateTaskModalProps) => {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const memberOptions = projectMembers.map((member) => ({
    label: member.user.name,
    value: member.user.id,
  }));

  const {
    control,
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useForm<ITaskForm>({
    resolver: zodResolver(taskSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createTask,
    onSuccess: () => {
      handleSuccess('Task created successfully');
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.all });
      setOpen(false);
      reset();
    },
  });

  const onSubmit: SubmitHandler<ITaskForm> = (form) => {
    mutate({
      projectId,
      form,
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button" className="flex w-fit items-center gap-2">
          <Plus className="h-4 w-4" />
          Create task
        </Button>
      </DialogTrigger>
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
                options={memberOptions}
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
          <TextEditor
            initialValue=""
            onChange={(html) => setValue('description', html)}
            label="Description"
            id="description"
            placeholder="Enter task description"
            error={errors.description?.message}
          />
          <div className="flex w-full items-center justify-end gap-4">
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                Cancel
              </Button>
            </DialogClose>
            <Button type="submit" disabled={isPending}>
              {isPending ? <LoadingSpinner size="1.5em" /> : 'Create task'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};

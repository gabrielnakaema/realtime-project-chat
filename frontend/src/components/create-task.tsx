import { taskSchema, type ITaskForm } from '@/schemas/task-schema';
import { taskQueryKeys } from '@/services/query-keys';
import { createTask } from '@/services/tasks';
import { handleSuccess } from '@/utils/handle-success';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { useForm, type SubmitHandler } from 'react-hook-form';
import { Button } from './button';
import { Input } from './input';
import { LoadingSpinner } from './loading';
import { Textarea } from './textarea';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog';

interface CreateTaskModalProps {
  projectId: string;
}

export const CreateTask = ({ projectId }: CreateTaskModalProps) => {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<ITaskForm>({
    resolver: zodResolver(taskSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createTask,
    onSuccess: () => {
      handleSuccess('Task created successfully');
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.listByProjectId(projectId) });
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
        <Button type="button" className="w-fit flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Create task
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create task</DialogTitle>
          <DialogDescription>Create a new task for the project</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Title"
            id="title"
            placeholder="Enter task title"
            {...register('title')}
            error={errors.title?.message}
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
          <div className="flex items-center gap-4 w-full justify-end">
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

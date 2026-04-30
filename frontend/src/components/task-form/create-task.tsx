import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog';
import { TaskFormFields } from './task-form-fields';
import type { SubmitHandler } from 'react-hook-form';
import type { Member, ProjectColumn } from '@/types/project';
import type { ITaskForm } from '@/schemas/task-schema';
import { handleSuccess } from '@/utils/handle-success';
import { createTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { taskSchema } from '@/schemas/task-schema';

interface CreateTaskModalProps {
  projectId: string;
  projectMembers: Member[];
  projectColumns: ProjectColumn[];
}

export const CreateTask = ({ projectId, projectMembers, projectColumns }: CreateTaskModalProps) => {
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
    defaultValues: {
      project_column_id: projectColumns[0]?.id,
    },
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
          <input type="hidden" value={projectColumns[0]?.id} {...register('project_column_id')} />
          <TaskFormFields
            control={control}
            register={register}
            setValue={setValue}
            errors={errors}
            memberOptions={memberOptions}
            descriptionInitialValue=""
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

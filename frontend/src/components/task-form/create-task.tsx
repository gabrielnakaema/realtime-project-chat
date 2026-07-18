import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';
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
import { getTaskMemberOptions } from './task-form-utils';
import type { ReactNode } from 'react';
import type { SubmitHandler } from 'react-hook-form';
import type { ITaskForm } from '@/schemas/task-schema';
import { useProjectMembers } from '@/hooks/use-project-members';
import { handleSuccess } from '@/utils/handle-success';
import { reconcileTask } from '@/hooks/task-board-cache';
import { createTask } from '@/services/tasks';
import { taskQueryKeys } from '@/services/query-keys';
import { taskSchema } from '@/schemas/task-schema';

interface CreateTaskModalProps {
  projectId: string;
  initialProjectColumnId?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  trigger?: ReactNode | null;
}

export const CreateTask = ({
  projectId,
  initialProjectColumnId,
  open: controlledOpen,
  onOpenChange,
  trigger,
}: CreateTaskModalProps) => {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false);
  const queryClient = useQueryClient();
  const open = controlledOpen ?? uncontrolledOpen;
  const setOpen = (nextOpen: boolean) => {
    if (controlledOpen === undefined) {
      setUncontrolledOpen(nextOpen);
    }

    onOpenChange?.(nextOpen);
  };
  const { data: projectMembers, isLoading: isProjectMembersLoading } = useProjectMembers(projectId, open);
  const memberOptions = getTaskMemberOptions(projectMembers ?? []);

  const form = useForm<ITaskForm>({
    resolver: zodResolver(taskSchema as any) as any,
    defaultValues: {
      project_column_id: initialProjectColumnId,
      depends_on_task_ids: [],
    },
  });
  const { reset } = form;

  useEffect(() => {
    if (!open) {
      return;
    }

    reset({
      project_column_id: initialProjectColumnId,
      depends_on_task_ids: [],
    });
  }, [initialProjectColumnId, open, reset]);

  const { mutate, isPending } = useMutation({
    mutationFn: createTask,
    onSuccess: (task) => {
      handleSuccess('Task created successfully');
      reconcileTask(queryClient, task);
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.listUserDueTasks });
      setOpen(false);
      reset({
        project_column_id: initialProjectColumnId,
        depends_on_task_ids: [],
      });
    },
  });

  const onSubmit: SubmitHandler<ITaskForm> = (data) => {
    mutate({
      projectId,
      form: data,
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      {trigger === undefined && (
        <DialogTrigger asChild>
          <Button type="button" className="flex w-fit items-center gap-2">
            <Plus className="h-4 w-4" />
            Create task
          </Button>
        </DialogTrigger>
      )}
      {trigger && <DialogTrigger asChild>{trigger}</DialogTrigger>}
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col gap-0 overflow-hidden p-0 md:max-w-5xl">
        <DialogHeader className="border-border shrink-0 border-b px-6 py-5">
          <DialogTitle>Create task</DialogTitle>
          <DialogDescription>Create a new task for the project</DialogDescription>
        </DialogHeader>

        <FormProvider {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="flex min-h-0 w-full flex-1 flex-col overflow-hidden">
            <div className="flex-1 overflow-y-auto px-6 py-5">
              <input type="hidden" {...form.register('project_column_id')} />
              {isProjectMembersLoading ? (
                <div className="flex min-h-50 items-center justify-center">
                  <LoadingSpinner size="3rem" />
                </div>
              ) : (
                <TaskFormFields projectId={projectId} memberOptions={memberOptions} descriptionInitialValue="" />
              )}
            </div>
            <div className="border-border flex w-full shrink-0 items-center justify-end gap-4 border-t px-6 py-4">
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
        </FormProvider>
      </DialogContent>
    </Dialog>
  );
};

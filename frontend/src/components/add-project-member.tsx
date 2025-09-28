import { Plus } from 'lucide-react';
import { zodResolver } from '@hookform/resolvers/zod';
import {  useForm } from 'react-hook-form';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog';
import { Input } from './input';
import { LoadingSpinner } from './loading';
import { Button } from './button';
import type {SubmitHandler} from 'react-hook-form';
import type {IProjectMemberForm} from '@/schemas/project-member-schema';
import {  projectMemberSchema } from '@/schemas/project-member-schema';
import { createProjectMember } from '@/services/projects';
import { handleSuccess } from '@/utils/handle-success';
import { projectQueryKeys } from '@/services/query-keys';

interface AddProjectMemberProps {
  projectId: string;
}

export const AddProjectMember = ({ projectId }: AddProjectMemberProps) => {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<IProjectMemberForm>({
    resolver: zodResolver(projectMemberSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createProjectMember,
    onSuccess: () => {
      handleSuccess('Member added successfully');
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      reset();
      setOpen(false);
    },
  });

  const onSubmit: SubmitHandler<IProjectMemberForm> = (form) => {
    mutate({
      projectId,
      email: form.email,
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button
          type="button"
          className="w-fit flex items-center gap-2 p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-md transition-colors"
        >
          <Plus className="w-6 h-6 text-slate-500" />
        </button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add project member</DialogTitle>
          <DialogDescription>Add a new member to the project</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="gap-8 flex flex-col">
          <Input
            label="Email"
            id="email"
            placeholder="Enter email address"
            error={errors.email?.message}
            {...register('email')}
          />
          <Button type="submit" disabled={isPending} className="w-full">
            {isPending ? <LoadingSpinner size="1.5em" /> : 'Add member'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
};

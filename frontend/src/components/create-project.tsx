import { Button } from './button';
import { DialogContent, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog';
import { Plus } from 'lucide-react';
import { Dialog } from './ui/dialog';
import { useState } from 'react';
import { projectSchema, type IProjectForm } from '@/schemas/project-schema';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm, type SubmitHandler } from 'react-hook-form';
import { handleSuccess } from '@/utils/handle-success';
import { createProject } from '@/services/projects';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { projectQueryKeys } from '@/services/query-keys';
import { Input } from './input';
import { Textarea } from './textarea';
import { LoadingSpinner } from './loading';

export const CreateProject = () => {
  const queryClient = useQueryClient();

  const [open, setOpen] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<IProjectForm>({
    resolver: zodResolver(projectSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      handleSuccess('Project created successfully');
      setOpen(false);
      reset();
    },
  });

  const onSubmit: SubmitHandler<IProjectForm> = (form) => {
    mutate(form);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button" className="w-fit flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Create project
        </Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create project</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            id="name"
            label="Name"
            placeholder="Enter project name"
            error={errors.name?.message}
            {...register('name')}
          />
          <Textarea
            id="description"
            label="Description"
            placeholder="Enter project description"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleSubmit(onSubmit)();
              }
            }}
            error={errors.description?.message}
            {...register('description')}
          />
          <Button type="submit" disabled={isPending} className="ml-auto">
            {isPending ? <LoadingSpinner size="1.5em" /> : 'Create project'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
};

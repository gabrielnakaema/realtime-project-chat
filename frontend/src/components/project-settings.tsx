import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { Settings } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog';
import { Input } from './input';
import { Textarea } from './textarea';
import { Button } from './button';
import { LoadingSpinner } from './loading';
import type { SubmitHandler } from 'react-hook-form';
import type { IProjectForm } from '@/schemas/project-schema';
import { projectQueryKeys } from '@/services/query-keys';
import { getProject, updateProject } from '@/services/projects';
import { projectSchema } from '@/schemas/project-schema';
import { handleSuccess } from '@/utils/handle-success';

interface ProjectSettingsProps {
  projectId: string;
}

export const ProjectSettings = ({ projectId }: ProjectSettingsProps) => {
  const queryClient = useQueryClient();

  const [open, setOpen] = useState(false);
  const { data, isLoading } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
    enabled: open,
  });

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<IProjectForm>({
    resolver: zodResolver(projectSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: updateProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      handleSuccess('Project created successfully');
      setOpen(false);
      reset();
    },
  });

  useEffect(() => {
    if (data) {
      reset({
        name: data.name,
        description: data.description,
      });
    }
  }, [data, reset]);

  const onSubmit: SubmitHandler<IProjectForm> = (form) => {
    mutate({
      description: form.description,
      name: form.name,
      id: projectId,
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button className="inline-flex items-center px-3 py-2 text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-600 rounded-md font-medium transition-colors">
          <Settings className="w-4 h-4 mr-2" />
          Settings
        </button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Project settings</DialogTitle>
        </DialogHeader>

        {isLoading && (
          <div className="min-h-50 flex flex-col items-center justify-center">
            <LoadingSpinner size="3rem" />
          </div>
        )}

        {!isLoading && (
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
              {isPending ? <LoadingSpinner size="1.5em" /> : 'Save changes'}
            </Button>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
};

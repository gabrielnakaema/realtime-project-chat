import { FolderKanban, Plus } from 'lucide-react';
import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { FormProvider } from 'react-hook-form';
import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '../ui/dialog';
import { ProjectFormFields } from './project-form-fields';
import { useProjectForm } from './use-project-form';
import type { SubmitHandler } from 'react-hook-form';
import type { IProjectForm } from '@/schemas/project-schema';
import { createProject } from '@/services/projects';
import { handleSuccess } from '@/utils/handle-success';
import { projectQueryKeys } from '@/services/query-keys';

interface CreateProjectProps {
  trigger?: React.ReactNode;
}

export const CreateProject = ({ trigger }: CreateProjectProps = {}) => {
  const queryClient = useQueryClient();

  const [open, setOpen] = useState(false);
  const { form, resetForm } = useProjectForm();

  const { mutate, isPending } = useMutation({
    mutationFn: createProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      handleSuccess('Project created successfully');
      setOpen(false);
      resetForm();
    },
  });

  const onSubmit: SubmitHandler<IProjectForm> = (values) => {
    mutate(values);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button type="button" className="flex w-fit items-center gap-2">
            <Plus className="h-4 w-4" />
            Create project
          </Button>
        )}
      </DialogTrigger>

      <DialogContent className="flex max-h-[calc(100vh-5rem)] flex-col gap-0 overflow-y-auto p-0 md:max-w-5xl">
        <DialogHeader className="bg-background border-border sticky top-0 z-10 shrink-0 border-b px-6 py-5">
          <DialogTitle className="flex items-center gap-2">
            <FolderKanban className="text-primary h-5 w-5" />
            Create project
          </DialogTitle>
          <DialogDescription>
            Set up the project details and the workflow your team will use on the board.
          </DialogDescription>
        </DialogHeader>

        <FormProvider {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5 px-6 py-5">
            <ProjectFormFields
              mode="create"
              descriptionInitialValue=""
              descriptionEditorKey="create-project-description"
            />

            <Button type="submit" disabled={isPending} className="ml-auto flex">
              {isPending ? <LoadingSpinner size="1.5em" /> : 'Create project'}
            </Button>
          </form>
        </FormProvider>
      </DialogContent>
    </Dialog>
  );
};

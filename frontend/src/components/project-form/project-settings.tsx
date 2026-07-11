import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { Settings } from 'lucide-react';
import { FormProvider } from 'react-hook-form';
import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '../ui/dialog';
import { ProjectFormFields } from './project-form-fields';
import { getProjectFormValues, restoreMissingProjectColumnIds } from './project-form-utils';
import { useProjectForm } from './use-project-form';
import type { SubmitHandler } from 'react-hook-form';
import type { IProjectForm } from '@/schemas/project-schema';
import { invalidateProjectBoardData } from '@/services/project-board-invalidation';
import { projectQueryKeys } from '@/services/query-keys';
import { getProject, updateProject } from '@/services/projects';
import { handleSuccess } from '@/utils/handle-success';

interface ProjectSettingsProps {
  projectId: string;
}

export const ProjectSettings = ({ projectId }: ProjectSettingsProps) => {
  const queryClient = useQueryClient();

  const [open, setOpen] = useState(false);
  const { form, resetForm } = useProjectForm();
  const { data, isLoading } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
    enabled: open,
  });

  const { mutate, isPending } = useMutation({
    mutationFn: updateProject,
    onSuccess: async () => {
      await invalidateProjectBoardData(queryClient);
      handleSuccess('Project saved successfully');
      setOpen(false);
    },
  });

  useEffect(() => {
    if (!data) {
      return;
    }

    resetForm(getProjectFormValues(data));
  }, [data, resetForm]);

  const onSubmit: SubmitHandler<IProjectForm> = (values) => {
    const columns = data
      ? restoreMissingProjectColumnIds(values.columns, data.columns, values.deleted_columns)
      : values.columns;

    mutate({
      description: values.description,
      name: values.name,
      repository_url: values.repository_url,
      repository_owner: values.repository_owner,
      repository_name: values.repository_name,
      default_branch: values.default_branch,
      branch_name_prefix: values.branch_name_prefix,
      id: projectId,
      columns,
      deleted_columns: values.deleted_columns,
    });
  };

  let content = (
    <div className="flex min-h-50 flex-1 flex-col items-center justify-center px-6 py-6">
      <p className="text-sm text-slate-500 dark:text-slate-400">Project not found</p>
    </div>
  );

  if (isLoading) {
    content = (
      <div className="flex min-h-50 flex-1 flex-col items-center justify-center px-6 py-6">
        <LoadingSpinner size="3rem" />
      </div>
    );
  }

  if (data) {
    content = (
      <FormProvider {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5 px-6 py-5">
          <ProjectFormFields
            mode="settings"
            descriptionInitialValue={data.description}
            descriptionEditorKey={data.id}
            originalColumns={data.columns}
          />

          <Button type="submit" disabled={isPending} className="ml-auto flex">
            {isPending ? <LoadingSpinner size="1.5em" /> : 'Save changes'}
          </Button>
        </form>
      </FormProvider>
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button className="inline-flex items-center rounded-md border border-slate-300 bg-white px-3 py-2 font-medium text-slate-700 transition-colors hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-300 dark:hover:bg-slate-600">
          <Settings className="mr-2 h-4 w-4" />
          Settings
        </button>
      </DialogTrigger>

      <DialogContent className="flex max-h-[calc(100vh-5rem)] flex-col gap-0 overflow-y-auto p-0 md:max-w-5xl">
        <DialogHeader className="bg-background sticky top-0 z-10 shrink-0 border-b border-slate-200 px-6 py-5 dark:border-slate-700">
          <DialogTitle>Project settings</DialogTitle>
          <DialogDescription>
            Update project details and refine the board flow without changing the rest of the workspace.
          </DialogDescription>
        </DialogHeader>

        {content}
      </DialogContent>
    </Dialog>
  );
};

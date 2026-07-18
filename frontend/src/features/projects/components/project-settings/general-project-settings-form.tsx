import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import type { SubmitHandler } from 'react-hook-form';
import type { GeneralProjectSettingsFormData } from '@/features/projects/schemas/general-project-settings.schema';
import type { Project } from '@/types/project';
import { LoadingSpinner } from '@/components/loading';
import { TextEditor } from '@/shared/components/text-editor';
import { generalProjectSettingsSchema } from '@/features/projects/schemas/general-project-settings.schema';
import { useProjectDetails } from '@/hooks/use-project-details';
import { invalidateProjectBoardData } from '@/services/project-board-invalidation';
import { updateProject } from '@/services/projects';
import { ConfirmationDialog } from '@/shared/components/confirmation-dialog';
import { Input } from '@/shared/components/input';
import { Button } from '@/shared/components/button';
import { handleError } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';

const getGeneralProjectSettingsFormValues = (project?: Project): GeneralProjectSettingsFormData => ({
  branch_name_prefix: project?.branch_name_prefix ?? '',
  default_branch: project?.default_branch ?? '',
  description: project?.description ?? '',
  name: project?.name ?? '',
  repository_name: project?.repository_name ?? '',
  repository_owner: project?.repository_owner ?? '',
  repository_url: project?.repository_url ?? '',
});

export const GeneralProjectSettingsForm = ({ projectId }: { projectId: string }) => {
  const queryClient = useQueryClient();
  const [isDeletingProject, setIsDeletingProject] = useState(false);
  const [editorKey, setEditorKey] = useState(0);

  const { data: project, isLoading } = useProjectDetails(projectId);

  const {
    getValues,
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors, isDirty },
  } = useForm<GeneralProjectSettingsFormData>({
    resolver: zodResolver(generalProjectSettingsSchema),
    values: getGeneralProjectSettingsFormValues(project),
  });

  const updateMutation = useMutation({
    mutationFn: updateProject,
    onSuccess: async (updatedProject) => {
      reset(getGeneralProjectSettingsFormValues(updatedProject));
      setEditorKey((current) => current + 1);
      await invalidateProjectBoardData(queryClient);
      handleSuccess('Project saved successfully');
    },
    onError: (error) => {
      void handleError(error);
    },
  });

  const onSubmit: SubmitHandler<GeneralProjectSettingsFormData> = (form) => {
    if (!project) return;

    updateMutation.mutate({
      id: projectId,
      ...form,
      columns: project.columns.map((column) => ({
        id: column.id,
        name: column.name,
        description: column.description,
        color: column.color,
        is_done_column: column.is_done_column,
      })),
      deleted_columns: [],
    });
  };

  const handleDiscard = () => {
    reset(getGeneralProjectSettingsFormValues(project));
    setEditorKey((prev) => prev + 1);
  };

  if (isLoading) {
    return <GeneralProjectSettingsFormSkeleton />;
  }

  if (!project) {
    return <p className="text-muted-foreground text-sm">Project not found.</p>;
  }

  return (
    <form className="flex w-full flex-col gap-4" id="general-project-settings" onSubmit={handleSubmit(onSubmit)}>
      <Input
        id="name"
        label="PROJECT NAME"
        placeholder="Enter project name"
        {...register('name')}
        error={errors.name?.message}
      />

      <TextEditor
        key={editorKey}
        initialValue={getValues('description')}
        onChange={(newValue) => {
          setValue('description', newValue, {
            shouldDirty: true,
          });
        }}
        label="PROJECT DESCRIPTION"
        placeholder="Enter project description"
        id="description"
        error={errors.description?.message}
      />

      <hr className="bg-card min-h-[1px] w-full" />

      <p className="text-foreground text-[11px] font-semibold tracking-wider">REPOSITORY SETTINGS</p>

      <div className="grid w-full grid-cols-2 gap-4">
        <Input
          id="repository_url"
          label="URL"
          placeholder="https://github.com/acme/project-chat-pubsub"
          {...register('repository_url')}
          error={errors.repository_url?.message}
        />
        <Input
          id="repository_owner"
          label="OWNER"
          placeholder="acme"
          {...register('repository_owner')}
          error={errors.repository_owner?.message}
        />
        <Input
          id="repository_name"
          label="NAME"
          placeholder="project-chat-pubsub"
          {...register('repository_name')}
          error={errors.repository_name?.message}
        />
        <Input
          id="default_branch"
          label="DEFAULT BRANCH"
          placeholder="main"
          {...register('default_branch')}
          error={errors.default_branch?.message}
        />
        <Input
          id="branch_name_prefix"
          label="BRANCH NAME PREFIX"
          placeholder="task/"
          {...register('branch_name_prefix')}
          error={errors.branch_name_prefix?.message}
        />
      </div>

      <hr className="bg-card min-h-[1px] w-full" />

      <p className="text-destructive text-[11px] font-semibold tracking-wider">DANGER ZONE</p>

      <div className="border-destructive/30 flex w-full items-center justify-between rounded-md border p-3">
        <div>
          <p className="text-foreground text-xs font-medium">Delete project</p>
          <p className="text-muted-foreground text-xs font-normal">
            Permanently delete project, tasks, comments and history.
          </p>
        </div>

        <Button type="button" variant="destructiveOutline" size="sm" onClick={() => setIsDeletingProject(true)}>
          Delete
        </Button>
      </div>
      {isDirty && (
        <section className="border-border bg-card fixed bottom-0 left-[50%] flex min-h-20 w-full max-w-5xl -translate-x-1/2 items-center justify-between rounded-t-md border border-b-0 px-8 shadow-md">
          <p className="text-muted-foreground text-sm">You have unsaved changes.</p>

          <div className="flex items-center gap-2">
            <Button type="button" variant="outline" onClick={handleDiscard} disabled={updateMutation.isPending}>
              Discard
            </Button>
            <Button type="submit" form="general-project-settings" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? <LoadingSpinner size="1.25rem" /> : 'Save changes'}
            </Button>
          </div>
        </section>
      )}

      <ConfirmationDialog
        open={isDeletingProject}
        onOpenChange={setIsDeletingProject}
        variant="destructive"
        title={`Delete ${project.name}?`}
        description="This permanently deletes the project, tasks, comments and history. This action cannot be undone."
        confirmLabel="Delete project"
        onConfirm={() => {
          setIsDeletingProject(false);
        }}
      />
    </form>
  );
};

const GeneralProjectSettingsFormSkeleton = () => {
  return (
    <div className="flex w-full animate-pulse flex-col gap-4" aria-hidden="true">
      <div className="flex w-full flex-col gap-1.5">
        <div className="bg-card h-3 w-24 rounded" />
        <div className="bg-card h-[39px] w-full rounded-md" />
      </div>

      <div className="flex w-full flex-col gap-1.5">
        <div className="bg-card h-3 w-32 rounded" />
        <div className="bg-card h-9 w-full rounded-t-md" />
        <div className="bg-card h-40 w-full rounded-b-md" />
      </div>

      <hr className="bg-card min-h-[1px] w-full" />

      <div className="bg-card h-3 w-36 rounded" />

      <div className="grid w-full grid-cols-2 gap-4">
        {Array.from({ length: 5 }).map((_, index) => (
          <div key={index} className="flex w-full flex-col gap-1.5">
            <div className="bg-card h-3 w-20 rounded" />
            <div className="bg-card h-[39px] w-full rounded-md" />
          </div>
        ))}
      </div>

      <hr className="bg-card min-h-[1px] w-full" />

      <div className="bg-card h-3 w-24 rounded" />

      <div className="border-border flex w-full items-center justify-between rounded-md border p-3">
        <div className="flex flex-col gap-1.5">
          <div className="bg-card h-3 w-28 rounded" />
          <div className="bg-card h-3 w-64 rounded" />
        </div>
        <div className="bg-card h-8 w-16 rounded-md" />
      </div>
    </div>
  );
};

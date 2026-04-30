import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { Settings } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog';
import { Input } from './input';
import { ProjectColumnsEditor } from './project-columns-editor';
import { TextEditor } from './text-editor';
import { Button } from './button';
import { LoadingSpinner } from './loading';
import type { SubmitHandler } from 'react-hook-form';
import type { IProjectForm } from '@/schemas/project-schema';
import { projectQueryKeys } from '@/services/query-keys';
import { getProject, updateProject } from '@/services/projects';
import { projectSchema } from '@/schemas/project-schema';
import { handleSuccess } from '@/utils/handle-success';
import { buildProjectColumnSurface } from '@/lib/project-column-colors';

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
    setValue,
    watch,
  } = useForm<IProjectForm>({
    resolver: zodResolver(projectSchema as any) as any,
    defaultValues: {
      columns: [],
      deleted_columns: [],
    },
  });

  const columns = watch('columns');
  const deletedColumns = watch('deleted_columns');

  const { mutate, isPending } = useMutation({
    mutationFn: updateProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      handleSuccess('Project saved successfully');
      setOpen(false);
      reset();
    },
  });

  useEffect(() => {
    if (data) {
      reset({
        name: data.name,
        description: data.description,
        columns: data.columns.map((column) => ({
          id: column.id,
          name: column.name,
          color: column.color,
          is_done_column: column.is_done_column,
        })),
        deleted_columns: [],
      });
    }
  }, [data, reset]);

  const onSubmit: SubmitHandler<IProjectForm> = (form) => {
    mutate({
      description: form.description,
      name: form.name,
      id: projectId,
      columns: form.columns,
      deleted_columns: form.deleted_columns,
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button className="inline-flex items-center rounded-md border border-slate-300 bg-white px-3 py-2 font-medium text-slate-700 transition-colors hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-300 dark:hover:bg-slate-600">
          <Settings className="mr-2 h-4 w-4" />
          Settings
        </button>
      </DialogTrigger>

      <DialogContent className="flex max-h-[calc(100vh-5rem)] flex-col gap-0 overflow-y-auto p-0 md:max-w-3xl">
        <DialogHeader className="bg-background sticky top-0 z-10 shrink-0 border-b border-slate-200 px-6 py-5 dark:border-slate-700">
          <DialogTitle>Project settings</DialogTitle>
          <DialogDescription>
            Update project details and refine the board flow without changing the rest of the workspace.
          </DialogDescription>
        </DialogHeader>

        {isLoading && (
          <div className="flex min-h-50 flex-1 flex-col items-center justify-center px-6 py-6">
            <LoadingSpinner size="3rem" />
          </div>
        )}

        {!isLoading && (
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5 px-6 py-5">
            <div className="grid gap-5 md:grid-cols-[1.1fr_0.9fr]">
              <div className="space-y-4 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-700 dark:bg-slate-950">
                <div className="space-y-1">
                  <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Project details</h3>
                  <p className="text-sm text-slate-600 dark:text-slate-400">
                    Keep the name and description aligned with how the team uses this space.
                  </p>
                </div>

                <Input
                  id="name"
                  label="Name"
                  placeholder="Enter project name"
                  error={errors.name?.message}
                  {...register('name')}
                />
                <TextEditor
                  initialValue={data?.description ?? ''}
                  onChange={(html) => setValue('description', html)}
                  label="Description"
                  id="description"
                  error={errors.description?.message}
                  placeholder="Enter project description"
                />
              </div>

              <div className="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-900/50">
                <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">Current workflow</p>
                <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
                  Reorder stages, rename them, or remove old steps. Removed columns will ask where their tasks should
                  move.
                </p>
                <div className="mt-4 flex flex-wrap gap-2">
                  {columns.map((column, index) => (
                    <span
                      key={column.id ?? `${column.name}-${index}`}
                      className="rounded-full border px-3 py-1 text-xs font-medium"
                      style={{
                        backgroundColor: buildProjectColumnSurface(column.color).badgeBackground,
                        borderColor: buildProjectColumnSurface(column.color).borderColor,
                        color: column.color,
                      }}
                    >
                      {column.name || `Column ${index + 1}`}
                      {column.is_done_column ? ' • done' : ''}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            <ProjectColumnsEditor
              mode="settings"
              columns={columns}
              onChange={(nextColumns) => setValue('columns', nextColumns)}
              deletedColumns={deletedColumns}
              onDeletedColumnsChange={(nextDeletedColumns) => setValue('deleted_columns', nextDeletedColumns)}
              originalColumns={data?.columns ?? []}
              error={errors.columns?.message}
            />

            <Button type="submit" disabled={isPending} className="ml-auto flex">
              {isPending ? <LoadingSpinner size="1.5em" /> : 'Save changes'}
            </Button>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
};

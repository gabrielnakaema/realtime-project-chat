import { FolderKanban, Plus } from 'lucide-react';
import { useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog';
import { Button } from './button';
import { Input } from './input';
import { LoadingSpinner } from './loading';
import { ProjectColumnsEditor } from './project-columns-editor';
import { TextEditor } from './text-editor';
import type { SubmitHandler } from 'react-hook-form';
import type { IProjectForm } from '@/schemas/project-schema';
import type { ProjectColumn } from '@/types/project';
import { createProject } from '@/services/projects';
import { projectSchema } from '@/schemas/project-schema';
import { handleSuccess } from '@/utils/handle-success';
import { projectQueryKeys } from '@/services/query-keys';
import { buildProjectColumnSurface, getDefaultProjectColumnColor } from '@/lib/project-column-colors';

type ProjectPreviewColumn = Pick<ProjectColumn, 'id' | 'name'>;

const defaultColumns = (): Pick<ProjectColumn, 'name' | 'color' | 'is_done_column'>[] => [
  { name: 'Pending', color: getDefaultProjectColumnColor(0), is_done_column: false },
  { name: 'Doing', color: getDefaultProjectColumnColor(1), is_done_column: false },
  { name: 'Done', color: getDefaultProjectColumnColor(2), is_done_column: true },
];

const getColumnKey = (column: ProjectPreviewColumn, index: number) => column.id ?? `new-column-${index}`;

export const CreateProject = () => {
  const queryClient = useQueryClient();

  const [open, setOpen] = useState(false);

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
      columns: defaultColumns(),
      deleted_columns: [],
    },
  });

  const columns = watch('columns');

  const { mutate, isPending } = useMutation({
    mutationFn: createProject,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      handleSuccess('Project created successfully');
      setOpen(false);
      reset();
      setValue('columns', defaultColumns());
      setValue('deleted_columns', []);
    },
  });

  const onSubmit: SubmitHandler<IProjectForm> = (form) => {
    mutate(form);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button" className="flex w-fit items-center gap-2">
          <Plus className="h-4 w-4" />
          Create project
        </Button>
      </DialogTrigger>

      <DialogContent className="flex max-h-[calc(100vh-5rem)] flex-col gap-0 overflow-y-auto p-0 md:max-w-3xl">
        <DialogHeader className="bg-background sticky top-0 z-10 shrink-0 border-b border-slate-200 px-6 py-5 dark:border-slate-700">
          <DialogTitle className="flex items-center gap-2">
            <FolderKanban className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            Create project
          </DialogTitle>
          <DialogDescription>
            Set up the project details and the workflow your team will use on the board.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5 px-6 py-5">
          <div className="grid gap-5 md:grid-cols-[1.1fr_0.9fr]">
            <div className="space-y-4 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-700 dark:bg-slate-950">
              <div className="space-y-1">
                <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Project details</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400">
                  Give the team enough context to know what this board is for.
                </p>
              </div>

              <Input
                id="name"
                label="Name"
                placeholder="Website redesign"
                error={errors.name?.message}
                {...register('name')}
              />

              <TextEditor
                initialValue=""
                onChange={(html) => setValue('description', html)}
                label="Description"
                id="description"
                error={errors.description?.message}
                placeholder="What is this project about?"
              />
            </div>

            <div className="rounded-2xl border border-blue-200 bg-blue-50/80 p-4 text-sm text-blue-950 shadow-sm dark:border-blue-900 dark:bg-blue-950/30 dark:text-blue-100">
              <p className="font-semibold">A good flow stays lightweight.</p>
              <p className="mt-2 text-blue-900/80 dark:text-blue-100/80">
                Start with the default three-column board if you want something simple, or add review and testing stages
                now.
              </p>
              <div className="mt-4 flex flex-wrap gap-2">
                {columns.map((column, index) => (
                  <span
                    key={getColumnKey(column, index)}
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
            mode="create"
            columns={columns}
            onChange={(nextColumns) => setValue('columns', nextColumns)}
            error={errors.columns?.message}
          />

          <Button type="submit" disabled={isPending} className="ml-auto flex">
            {isPending ? <LoadingSpinner size="1.5em" /> : 'Create project'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
};

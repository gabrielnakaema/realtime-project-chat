import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowDown, ArrowUp, Check, CircleCheck, Plus, RotateCcw, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { useFieldArray, useForm, useWatch } from 'react-hook-form';
import type { SubmitHandler } from 'react-hook-form';
import type { ColumnsProjectSettingsFormData } from '@/features/projects/schemas/columns-project-settings.schema';
import type { Project, ProjectColumn } from '@/types/project';
import { LoadingSpinner } from '@/components/loading';
import { Select } from '@/components/select';
import { columnsProjectSettingsSchema } from '@/features/projects/schemas/columns-project-settings.schema';
import { buildProjectColumnSurface, getDefaultProjectColumnColor } from '@/lib/project-column-colors';
import { cn } from '@/lib/utils';
import { useProjectDetails } from '@/hooks/use-project-details';
import { invalidateProjectBoardData } from '@/services/project-board-invalidation';
import { updateProject } from '@/services/projects';
import { Button } from '@/shared/components/button';
import { Input } from '@/shared/components/input';
import { Input as Textarea } from '@/shared/components/textarea';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { handleError } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';

const getColumnsFormValues = (project?: Project): ColumnsProjectSettingsFormData => ({
  columns:
    project?.columns.map((column) => ({
      id: column.id,
      name: column.name,
      description: column.description,
      color: column.color,
      is_done_column: column.is_done_column,
    })) ?? [],
  deleted_columns: [],
});

const ensureDoneColumn = (columns: ColumnsProjectSettingsFormData['columns']) => {
  if (columns.length === 0 || columns.some((column) => column.is_done_column)) {
    return columns;
  }

  return columns.map((column, index) => ({
    ...column,
    is_done_column: index === columns.length - 1,
  }));
};

export const ColumnsProjectSettings = ({ projectId }: { projectId: string }) => {
  const queryClient = useQueryClient();
  const { data: project, isLoading } = useProjectDetails(projectId);
  const [openColumns, setOpenColumns] = useState<string[]>(['column-0']);

  const {
    control,
    formState: { errors, isDirty },
    handleSubmit,
    register,
    reset,
    setValue,
  } = useForm<ColumnsProjectSettingsFormData>({
    resolver: zodResolver(columnsProjectSettingsSchema),
    values: getColumnsFormValues(project),
  });

  const { append, fields, move, remove } = useFieldArray({
    control,
    name: 'columns',
    keyName: 'fieldId',
  });
  const columns = useWatch({ control, name: 'columns' });
  const deletedColumns = useWatch({ control, name: 'deleted_columns' });

  const updateMutation = useMutation({
    mutationFn: updateProject,
    onSuccess: async (updatedProject) => {
      reset(getColumnsFormValues(updatedProject));
      await invalidateProjectBoardData(queryClient);
      handleSuccess('Project columns saved successfully');
    },
    onError: (error) => {
      void handleError(error);
    },
  });

  const handleAddColumn = () => {
    const newColumnIndex = fields.length;

    append(
      {
        name: '',
        description: '',
        color: getDefaultProjectColumnColor(newColumnIndex),
        is_done_column: false,
      },
      { shouldFocus: false },
    );
    setOpenColumns((current) => [...current, `column-${newColumnIndex}`]);
  };

  const handleSetDoneColumn = (columnIndex: number) => {
    setValue(
      'columns',
      columns.map((column, index) => ({
        ...column,
        is_done_column: index === columnIndex,
      })),
      { shouldDirty: true, shouldValidate: true },
    );
  };

  const handleMoveColumn = (columnIndex: number, nextIndex: number) => {
    setOpenColumns((current) =>
      current.map((item) => {
        if (item === `column-${columnIndex}`) return `column-${nextIndex}`;
        if (item === `column-${nextIndex}`) return `column-${columnIndex}`;
        return item;
      }),
    );
    move(columnIndex, nextIndex);
  };

  const handleRemoveColumn = (columnIndex: number) => {
    const column = columns[columnIndex];
    const nextColumns = columns.filter((_, index) => index !== columnIndex);

    if (column.id) {
      const reassignmentTarget = nextColumns.find((candidate) => candidate.id);
      if (!reassignmentTarget?.id) return;

      setValue(
        'deleted_columns',
        [
          ...deletedColumns.filter((deletedColumn) => deletedColumn.id !== column.id),
          {
            id: column.id,
            move_tasks_to_column_id: reassignmentTarget.id,
          },
        ],
        { shouldDirty: true },
      );
    }

    remove(columnIndex);

    if (column.is_done_column) {
      setValue('columns', ensureDoneColumn(nextColumns), {
        shouldDirty: true,
        shouldValidate: true,
      });
    }
  };

  const handleRestoreColumn = (column: ProjectColumn) => {
    setValue(
      'deleted_columns',
      deletedColumns.filter((deletedColumn) => deletedColumn.id !== column.id),
      { shouldDirty: true },
    );
    append({
      id: column.id,
      name: column.name,
      description: column.description,
      color: column.color,
      is_done_column: !columns.some((candidate) => candidate.is_done_column) && column.is_done_column,
    });
  };

  const handleReassignmentChange = (columnId: string, targetColumnId: string) => {
    setValue(
      'deleted_columns',
      deletedColumns.map((deletedColumn) =>
        deletedColumn.id === columnId ? { ...deletedColumn, move_tasks_to_column_id: targetColumnId } : deletedColumn,
      ),
      { shouldDirty: true, shouldValidate: true },
    );
  };

  const handleDiscard = () => {
    reset(getColumnsFormValues(project));
    setOpenColumns(['column-0']);
  };

  const onSubmit: SubmitHandler<ColumnsProjectSettingsFormData> = (form) => {
    if (!project) return;

    updateMutation.mutate({
      id: projectId,
      name: project.name,
      description: project.description,
      repository_url: project.repository_url,
      repository_owner: project.repository_owner,
      repository_name: project.repository_name,
      default_branch: project.default_branch,
      branch_name_prefix: project.branch_name_prefix,
      columns: form.columns,
      deleted_columns: form.deleted_columns,
    });
  };

  if (isLoading) {
    return <ColumnsProjectSettingsSkeleton />;
  }

  if (!project) {
    return <p className="text-muted-foreground text-sm">Project not found.</p>;
  }

  const removedColumns = project.columns.filter((column) =>
    deletedColumns.some((deletedColumn) => deletedColumn.id === column.id),
  );
  const existingColumnCount = columns.filter((column) => column.id).length;
  const columnError = errors.columns?.message;

  return (
    <form className="flex w-full flex-col gap-4 pb-48" id="columns-project-settings" onSubmit={handleSubmit(onSubmit)}>
      <div className="flex w-full items-center justify-between gap-4">
        <div>
          <p className="text-foreground text-xs font-semibold tracking-wider">COLUMNS &bull; {columns.length}</p>
          <p className="text-muted-foreground mt-1 text-xs">Customize the stages tasks move through on this project.</p>
        </div>
        <Button type="button" onClick={handleAddColumn}>
          <Plus className="size-4" />
          Add column
        </Button>
      </div>

      <Accordion type="multiple" value={openColumns} onValueChange={setOpenColumns} className="space-y-2">
        {fields.map((field, index) => {
          const column = columns[index] ?? field;
          const columnLabel = column.name.trim() || `Untitled column ${index + 1}`;
          const surface = buildProjectColumnSurface(column.color);
          const canDelete = columns.length > 1 && (!column.id || existingColumnCount > 1);

          return (
            <AccordionItem
              key={field.fieldId}
              value={`column-${index}`}
              className="border-border bg-background overflow-hidden rounded-md border px-4 shadow-sm last:border-b"
            >
              <AccordionTrigger
                className="hover:no-underline"
                actions={
                  <div className="flex shrink-0 flex-col items-center">
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-7"
                      aria-label={`Move ${columnLabel} up`}
                      disabled={index === 0}
                      onClick={() => handleMoveColumn(index, index - 1)}
                    >
                      <ArrowUp className="size-3.5" />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="size-7"
                      aria-label={`Move ${columnLabel} down`}
                      disabled={index === columns.length - 1}
                      onClick={() => handleMoveColumn(index, index + 1)}
                    >
                      <ArrowDown className="size-3.5" />
                    </Button>
                  </div>
                }
              >
                <div className="flex min-w-0 items-center gap-3">
                  <span
                    className="size-2.5 shrink-0 rounded-full"
                    style={{ backgroundColor: surface.accentColor }}
                    aria-hidden="true"
                  />
                  <span className="min-w-0">
                    <span className="text-foreground block truncate text-[13px] font-semibold">{columnLabel}</span>
                    <span className="text-muted-foreground mt-0.5 block text-xs font-normal">
                      {column.description.trim() || 'No instructions added'}
                    </span>
                  </span>
                  {column.is_done_column && (
                    <span className="border-success/30 bg-success/10 text-success inline-flex shrink-0 items-center gap-1 rounded-full border px-2 py-0.5 text-[10px] font-semibold tracking-wide uppercase">
                      <CircleCheck className="size-3" />
                      Done
                    </span>
                  )}
                </div>
              </AccordionTrigger>

              <AccordionContent className="border-border border-t pt-4">
                <div className="flex flex-col gap-4">
                  <div className="grid gap-4 sm:grid-cols-[minmax(0,1fr)_140px]">
                    <Input
                      id={`column-name-${index}`}
                      label="COLUMN NAME"
                      placeholder="Enter column name"
                      error={errors.columns?.[index]?.name?.message}
                      {...register(`columns.${index}.name`)}
                    />

                    <div className="flex w-full flex-col gap-1.5">
                      <label htmlFor={`column-color-${index}`} className="text-[11px] font-semibold tracking-wider">
                        COLOR
                      </label>
                      <label
                        htmlFor={`column-color-${index}`}
                        className="border-border flex h-[39px] cursor-pointer items-center gap-2 rounded-md border bg-slate-900 px-3"
                      >
                        <input
                          id={`column-color-${index}`}
                          type="color"
                          className="size-5 cursor-pointer rounded border-0 bg-slate-900 p-0"
                          aria-label={`Color for ${columnLabel}`}
                          {...register(`columns.${index}.color`)}
                        />
                        <span className="text-muted-foreground font-mono text-xs uppercase">{column.color}</span>
                      </label>
                      {errors.columns?.[index]?.color?.message && (
                        <span className="text-destructive text-xs" role="alert">
                          {errors.columns[index].color.message}
                        </span>
                      )}
                    </div>
                  </div>

                  <Textarea
                    id={`column-description-${index}`}
                    label="INSTRUCTIONS"
                    placeholder="Describe what belongs in this column"
                    error={errors.columns?.[index]?.description?.message}
                    classNames={{
                      input: 'min-h-24 resize-y py-2',
                      error: 'text-destructive text-xs',
                    }}
                    {...register(`columns.${index}.description`)}
                  />

                  <div className="border-border flex flex-col justify-between gap-3 border-t pt-4 sm:flex-row sm:items-center">
                    <button
                      type="button"
                      className={cn(
                        'focus-visible:ring-ring/50 flex items-center gap-3 rounded-md text-left outline-none focus-visible:ring-3',
                        column.is_done_column ? 'text-success' : 'text-muted-foreground hover:text-foreground',
                      )}
                      onClick={() => handleSetDoneColumn(index)}
                      aria-pressed={column.is_done_column}
                    >
                      <span
                        className={cn(
                          'flex size-5 items-center justify-center rounded-full border',
                          column.is_done_column
                            ? 'border-success bg-success text-white'
                            : 'border-border bg-background',
                        )}
                      >
                        {column.is_done_column && <Check className="size-3.5" />}
                      </span>
                      <span>
                        <span className="text-foreground block text-xs font-medium">Done column</span>
                        <span className="text-muted-foreground block text-[11px]">
                          Completing a task moves it to this column.
                        </span>
                      </span>
                    </button>

                    <div className="flex items-center gap-2 self-end sm:self-auto">
                      <Button
                        type="button"
                        variant="destructiveOutline"
                        size="icon"
                        aria-label={`Delete ${columnLabel}`}
                        disabled={!canDelete}
                        title={!canDelete ? 'Keep another saved column available for task reassignment.' : undefined}
                        onClick={() => handleRemoveColumn(index)}
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          );
        })}
      </Accordion>

      {columnError && (
        <p className="text-destructive text-xs" role="alert">
          {columnError}
        </p>
      )}

      {removedColumns.length > 0 && (
        <section className="border-destructive/30 bg-destructive/5 flex flex-col gap-3 rounded-md border p-4">
          <div>
            <p className="text-destructive text-xs font-semibold tracking-wider">PENDING REMOVALS</p>
            <p className="text-muted-foreground mt-1 text-xs">Choose where existing tasks should move when you save.</p>
          </div>

          {removedColumns.map((removedColumn) => {
            const deletion = deletedColumns.find((deletedColumn) => deletedColumn.id === removedColumn.id);
            const reassignmentOptions = columns.flatMap((column) =>
              column.id
                ? [
                    {
                      label: column.name,
                      value: column.id,
                    },
                  ]
                : [],
            );

            return (
              <div
                key={removedColumn.id}
                className="border-destructive/20 bg-background flex flex-col gap-3 rounded-md border p-3 sm:flex-row sm:items-start"
              >
                <div className="flex min-w-0 flex-1 flex-col gap-2">
                  <p className="text-muted-foreground text-xs">
                    <strong>{removedColumn.name}</strong> column will be deleted after tasks are moved.
                  </p>
                  <Select
                    id={`reassign-${removedColumn.id}`}
                    label="MOVE TASKS TO"
                    value={deletion?.move_tasks_to_column_id ?? ''}
                    onChange={(value) => handleReassignmentChange(removedColumn.id, value)}
                    options={reassignmentOptions}
                    placeholder="Choose a column"
                  />
                </div>

                <Button type="button" variant="outline" size="sm" onClick={() => handleRestoreColumn(removedColumn)}>
                  <RotateCcw className="size-3.5" />
                  Restore
                </Button>
              </div>
            );
          })}
        </section>
      )}

      {isDirty && (
        <section className="border-border bg-card fixed bottom-0 left-1/2 z-20 flex min-h-20 w-full max-w-5xl -translate-x-1/2 items-center justify-between gap-4 rounded-t-md border border-b-0 px-4 shadow-md sm:px-8">
          <p className="text-muted-foreground text-sm">You have unsaved changes.</p>
          <div className="flex items-center gap-2">
            <Button type="button" variant="outline" onClick={handleDiscard} disabled={updateMutation.isPending}>
              Discard
            </Button>
            <Button type="submit" form="columns-project-settings" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? <LoadingSpinner size="1.25rem" /> : 'Save changes'}
            </Button>
          </div>
        </section>
      )}
    </form>
  );
};

const ColumnsProjectSettingsSkeleton = () => (
  <div className="flex w-full animate-pulse flex-col gap-4" aria-label="Loading project columns">
    <div className="flex items-center justify-between">
      <div className="space-y-2">
        <div className="bg-card h-3 w-24 rounded" />
        <div className="bg-card h-3 w-72 rounded" />
      </div>
      <div className="bg-card h-9 w-28 rounded-md" />
    </div>
    {Array.from({ length: 3 }).map((_, index) => (
      <div key={index} className="border-border bg-card h-16 rounded-md border" />
    ))}
  </div>
);

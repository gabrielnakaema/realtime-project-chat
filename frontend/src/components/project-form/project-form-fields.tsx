import { useFormContext } from 'react-hook-form';
import { TextEditor } from '../text-editor';
import { Input } from '../input';
import { getProjectColumnKey } from './project-form-utils';
import { ProjectColumnsEditor } from './project-columns-editor';
import type { IProjectForm } from '@/schemas/project-schema';
import type { ProjectColumn } from '@/types/project';
import { buildProjectColumnSurface } from '@/lib/project-column-colors';

interface ProjectFormFieldsProps {
  mode: 'create' | 'settings';
  descriptionInitialValue: string;
  descriptionEditorKey?: string;
  originalColumns?: ProjectColumn[];
}

export const ProjectFormFields = ({
  mode,
  descriptionInitialValue,
  descriptionEditorKey,
  originalColumns = [],
}: ProjectFormFieldsProps) => {
  const {
    register,
    setValue,
    watch,
    formState: { errors },
  } = useFormContext<IProjectForm>();

  const columns = watch('columns');
  const deletedColumns = watch('deleted_columns');

  const detailsCopy =
    mode === 'create'
      ? 'Give the team enough context to know what this board is for.'
      : 'Keep the name and description aligned with how the team uses this space.';
  const descriptionPlaceholder = mode === 'create' ? 'What is this project about?' : 'Enter project description';
  const namePlaceholder = mode === 'create' ? 'Website redesign' : 'Enter project name';
  const previewClassName =
    mode === 'create'
      ? 'rounded-2xl border border-blue-200 bg-blue-50/80 p-4 text-sm text-blue-950 shadow-sm dark:border-blue-900 dark:bg-blue-950/30 dark:text-blue-100'
      : 'rounded-2xl border border-slate-200 bg-slate-50/80 p-4 shadow-sm dark:border-slate-700 dark:bg-slate-900/50';
  const previewTitle = mode === 'create' ? 'A good flow stays lightweight.' : 'Current workflow';
  const previewDescription =
    mode === 'create'
      ? 'Start with the default three-column board if you want something simple, or add review and testing stages now.'
      : 'Reorder stages, rename them, or remove old steps. Removed columns will ask where their tasks should move.';
  const previewDescriptionClassName =
    mode === 'create'
      ? 'mt-2 text-blue-900/80 dark:text-blue-100/80'
      : 'mt-1 text-sm text-slate-600 dark:text-slate-400';

  return (
    <>
      <div className="">
        <div className="space-y-4 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-700 dark:bg-slate-950">
          <div className="space-y-1">
            <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Project details</h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">{detailsCopy}</p>
          </div>

          <Input
            id="name"
            label="Name"
            placeholder={namePlaceholder}
            error={errors.name?.message}
            {...register('name')}
          />

          <TextEditor
            key={descriptionEditorKey}
            initialValue={descriptionInitialValue}
            onChange={(html) => setValue('description', html)}
            label="Description"
            id="description"
            error={errors.description?.message}
            placeholder={descriptionPlaceholder}
          />
        </div>
      </div>
      <div className={previewClassName}>
        <p className="font-semibold">{previewTitle}</p>
        <p className={previewDescriptionClassName}>{previewDescription}</p>
        <div className="mt-4 flex flex-wrap gap-2">
          {columns.map((column, index) => {
            const surface = buildProjectColumnSurface(column.color);

            return (
              <span
                key={getProjectColumnKey(column, index)}
                className="rounded-full border px-3 py-1 text-xs font-medium"
                style={{
                  backgroundColor: surface.badgeBackground,
                  borderColor: surface.borderColor,
                  color: column.color,
                }}
              >
                {column.name || `Column ${index + 1}`}
                {column.is_done_column ? ' • done' : ''}
              </span>
            );
          })}
        </div>
      </div>

      <ProjectColumnsEditor
        mode={mode}
        columns={columns}
        onChange={(nextColumns) => setValue('columns', nextColumns)}
        deletedColumns={deletedColumns}
        onDeletedColumnsChange={(nextDeletedColumns) => setValue('deleted_columns', nextDeletedColumns)}
        originalColumns={originalColumns}
        error={errors.columns?.message}
      />
    </>
  );
};

import { useFormContext } from 'react-hook-form';
import { getProjectColumnKey } from './project-form-utils';
import { ProjectColumnsEditor } from './project-columns-editor';
import type { IProjectForm } from '@/features/projects/schemas/project.schema';
import type { ProjectColumn } from '@/features/projects/types/project';
import { Input } from '@/shared/components/input';
import { TextEditor } from '@/shared/components/text-editor';
import { buildProjectColumnSurface } from '@/features/projects/utils/project-column-colors';

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
      ? 'rounded-2xl border border-primary/30 bg-primary/10 p-4 text-sm text-primary shadow-sm   '
      : 'rounded-2xl border border-border bg-muted/80 p-4 shadow-sm  ';
  const previewTitle = mode === 'create' ? 'A good flow stays lightweight.' : 'Current workflow';
  const previewDescription =
    mode === 'create'
      ? 'Start with the default three-column board if you want something simple, or add review and testing stages now.'
      : 'Reorder stages, rename them, or remove old steps. Removed columns will ask where their tasks should move.';
  const previewDescriptionClassName = mode === 'create' ? 'mt-2 text-primary ' : 'mt-1 text-sm text-muted-foreground';

  return (
    <>
      <div>
        <div className="border-border bg-card space-y-4 rounded-2xl border p-4 shadow-sm">
          <div className="space-y-1">
            <h3 className="text-foreground text-sm font-semibold">Project details</h3>
            <p className="text-muted-foreground text-sm">{detailsCopy}</p>
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

          <div className="space-y-1 pt-2">
            <h3 className="text-foreground text-sm font-semibold">Repository details</h3>
            <p className="text-muted-foreground text-sm">
              Keep these optional fields filled in when this board maps to a code repository.
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <Input
              id="repository_url"
              label="Repository URL"
              placeholder="https://github.com/acme/project-chat-pubsub"
              error={errors.repository_url?.message}
              {...register('repository_url')}
            />

            <Input
              id="repository_owner"
              label="Repository Owner"
              placeholder="acme"
              error={errors.repository_owner?.message}
              {...register('repository_owner')}
            />

            <Input
              id="repository_name"
              label="Repository Name"
              placeholder="project-chat-pubsub"
              error={errors.repository_name?.message}
              {...register('repository_name')}
            />

            <Input
              id="default_branch"
              label="Default Branch"
              placeholder="main"
              error={errors.default_branch?.message}
              {...register('default_branch')}
            />

            <Input
              id="branch_name_prefix"
              label="Branch Name Prefix"
              placeholder="task/"
              error={errors.branch_name_prefix?.message}
              {...register('branch_name_prefix')}
            />
          </div>
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

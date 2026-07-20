import { ArrowDown, ArrowUp, CheckCircle2, Flag, Plus, RotateCcw, Trash2 } from 'lucide-react';
import { createDefaultProjectColumns } from './project-form-utils';
import type { IProjectForm } from '@/features/projects/schemas/project.schema';
import type { ProjectColumn } from '@/features/projects/types/project';
import { Button } from '@/components/button';
import { Input } from '@/components/input';
import { Select } from '@/shared/components/select';
import { Textarea } from '@/components/textarea';
import { cn } from '@/lib/utils';
import {
  buildProjectColumnSurface,
  getDefaultProjectColumnColor,
} from '@/features/projects/utils/project-column-colors';

type FormColumn = IProjectForm['columns'][number];
type DeletedColumn = IProjectForm['deleted_columns'][number];
type ColumnKeyItem = Pick<FormColumn, 'id'>;

interface ProjectColumnsEditorProps {
  columns: FormColumn[];
  onChange: (columns: FormColumn[]) => void;
  deletedColumns?: DeletedColumn[];
  onDeletedColumnsChange?: (columns: DeletedColumn[]) => void;
  originalColumns?: ProjectColumn[];
  mode: 'create' | 'settings';
  error?: string;
}

const ensureSingleDoneColumn = (columns: FormColumn[], preferredIndex = 0) =>
  columns.map((column, index) => ({
    ...column,
    is_done_column: index === preferredIndex,
  }));

const getColumnKey = (column: ColumnKeyItem, index: number) => column.id ?? `new-column-${index}`;

export const ProjectColumnsEditor = ({
  columns,
  onChange,
  deletedColumns = [],
  onDeletedColumnsChange,
  originalColumns = [],
  mode,
  error,
}: ProjectColumnsEditorProps) => {
  const existingColumns = columns.filter((column) => column.id);
  const removedColumns = originalColumns.filter((column) => deletedColumns.some((item) => item.id === column.id));

  const updateColumns = (nextColumns: FormColumn[]) => {
    const hasDoneColumn = nextColumns.some((column) => column.is_done_column);

    if (!hasDoneColumn && nextColumns.length > 0) {
      onChange(
        nextColumns.map((column, index) => ({
          ...column,
          is_done_column: index === nextColumns.length - 1,
        })),
      );
      return;
    }

    onChange(nextColumns);
  };

  const updateColumn = (index: number, updates: Partial<FormColumn>) => {
    updateColumns(columns.map((column, columnIndex) => (columnIndex === index ? { ...column, ...updates } : column)));
  };

  const moveColumn = (index: number, direction: -1 | 1) => {
    const nextIndex = index + direction;
    if (nextIndex < 0 || nextIndex >= columns.length) return;

    const nextColumns = [...columns];
    const [moved] = nextColumns.splice(index, 1);
    nextColumns.splice(nextIndex, 0, moved);
    updateColumns(nextColumns);
  };

  const setDoneColumn = (index: number) => {
    onChange(ensureSingleDoneColumn(columns, index));
  };

  const removeColumn = (index: number) => {
    const column = columns[index];
    const nextColumns = columns.filter((_, columnIndex) => columnIndex !== index);

    if (column.id && onDeletedColumnsChange) {
      const existingFallback = nextColumns.find((item) => item.id);
      if (existingFallback?.id) {
        const nextDeletedColumns = deletedColumns.filter((item) => item.id !== column.id);
        nextDeletedColumns.push({
          id: column.id,
          move_tasks_to_column_id: existingFallback.id,
        });
        onDeletedColumnsChange(nextDeletedColumns);
      }
    }

    updateColumns(nextColumns);
  };

  const restoreRemovedColumn = (removedColumn: ProjectColumn) => {
    onDeletedColumnsChange?.(deletedColumns.filter((item) => item.id !== removedColumn.id));
    updateColumns([
      ...columns,
      {
        id: removedColumn.id,
        name: removedColumn.name,
        description: removedColumn.description,
        color: removedColumn.color,
        is_done_column: removedColumn.is_done_column,
      },
    ]);
  };

  const availableReassignmentTargets = (removedId: string) =>
    columns
      .filter((column) => column.id && column.id !== removedId)
      .map((column) => ({
        label: column.name,
        value: column.id!,
      }));

  return (
    <section className="border-border bg-muted/80 space-y-4 rounded-2xl border p-4">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Flag className="text-primary h-4 w-4" />
            <h3 className="text-foreground text-sm font-semibold">Board columns</h3>
          </div>
          <p className="text-muted-foreground text-sm">
            Define the flow for this project. Every project needs at least one column and exactly one done column.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {mode === 'create' && (
            <Button
              type="button"
              variant="secondary"
              className="px-3 py-2 text-sm"
              onClick={() => onChange(createDefaultProjectColumns())}
            >
              <RotateCcw className="h-4 w-4" />
              Use default flow
            </Button>
          )}
          <Button
            type="button"
            variant="secondary"
            className="px-3 py-2 text-sm whitespace-nowrap"
            onClick={() =>
              onChange([
                ...columns,
                {
                  name: '',
                  description: '',
                  color: getDefaultProjectColumnColor(columns.length),
                  is_done_column: false,
                },
              ])
            }
          >
            <Plus className="h-4 w-4" />
            Add column
          </Button>
        </div>
      </div>

      <div className="grid gap-3">
        {columns.map((column, index) => {
          const canDelete = columns.length > 1 && (!column.id || existingColumns.length > 1);
          const surface = buildProjectColumnSurface(column.color);
          const columnLabel = column.name || `Column ${index + 1}`;

          return (
            <div
              key={getColumnKey(column, index)}
              style={{ borderColor: column.is_done_column ? undefined : surface.borderColor }}
              className={cn(
                'bg-card rounded-2xl border p-4 shadow-sm transition-colors',
                column.is_done_column ? 'border-success/30 ring-success/30 ring-1' : 'border-border',
              )}
            >
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start">
                <div className="flex min-w-0 flex-1 items-start gap-3">
                  <div className="bg-muted text-foreground flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-sm font-semibold">
                    {index + 1}
                  </div>

                  <div className="min-w-0 flex-1">
                    <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_150px]">
                      <Input
                        id={`column-${index}`}
                        label="Column name"
                        placeholder={index === 0 ? 'Backlog' : 'Column name'}
                        value={column.name}
                        onChange={(event) => updateColumn(index, { name: event.target.value })}
                      />

                      <div className="space-y-2">
                        <label htmlFor={`column-color-${index}`} className="text-foreground block text-sm font-medium">
                          Color
                        </label>
                        <div
                          className="flex items-center gap-3 rounded-xl border px-3 py-2"
                          style={{ backgroundColor: surface.backgroundColor, borderColor: surface.borderColor }}
                        >
                          <input
                            id={`column-color-${index}`}
                            type="color"
                            value={column.color}
                            onChange={(event) => updateColumn(index, { color: event.target.value })}
                            className="h-9 w-12 cursor-pointer rounded border-0 bg-transparent p-0"
                          />
                          <div className="min-w-0">
                            <p className="text-muted-foreground text-xs tracking-wide uppercase">Main color</p>
                            <p className="text-foreground font-mono text-sm">{column.color}</p>
                          </div>
                        </div>
                      </div>
                    </div>

                    <Textarea
                      id={`column-description-${index}`}
                      label="Column description"
                      rows={3}
                      value={column.description}
                      onChange={(event) => updateColumn(index, { description: event.target.value })}
                      placeholder="Optional guidance or instructions for this column."
                    />
                    <p className="text-muted-foreground text-xs">Optional.</p>

                    <div className="mt-3 flex flex-wrap items-center gap-2">
                      <span
                        className="inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm font-medium"
                        style={{
                          backgroundColor: surface.badgeBackground,
                          borderColor: surface.borderColor,
                          color: surface.accentColor,
                        }}
                      >
                        <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: surface.accentColor }} />
                        {column.name || 'Preview'}
                      </span>
                      <button
                        type="button"
                        onClick={() => setDoneColumn(index)}
                        className={cn(
                          'inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm font-medium transition-colors',
                          column.is_done_column
                            ? 'border-success/30 bg-success/10 text-success'
                            : 'border-border bg-card text-muted-foreground hover:border-border hover:text-foreground',
                        )}
                      >
                        <CheckCircle2 className="h-4 w-4" />
                        {column.is_done_column ? 'Done column' : 'Mark as done'}
                      </button>

                      {column.id && (
                        <span className="bg-muted text-muted-foreground rounded-full px-3 py-1 text-xs font-medium">
                          Existing column
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2 lg:pt-7">
                  <Button
                    type="button"
                    variant="secondary"
                    className="px-3 py-2"
                    aria-label={`Move ${columnLabel} up`}
                    onClick={() => moveColumn(index, -1)}
                    disabled={index === 0}
                  >
                    <ArrowUp className="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    className="px-3 py-2"
                    aria-label={`Move ${columnLabel} down`}
                    onClick={() => moveColumn(index, 1)}
                    disabled={index === columns.length - 1}
                  >
                    <ArrowDown className="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    className="px-3 py-2"
                    aria-label={`Delete ${columnLabel}`}
                    onClick={() => removeColumn(index)}
                    disabled={!canDelete}
                    title={!canDelete ? 'Keep at least one saved column available for task reassignment.' : undefined}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {removedColumns.length > 0 && onDeletedColumnsChange && (
        <div className="border-warning/30 bg-warning/10 space-y-3 rounded-2xl border p-4">
          <div className="space-y-1">
            <h4 className="text-warning text-sm font-semibold">Pending removals</h4>
            <p className="text-warning text-sm">Choose where tasks should move before these columns are removed.</p>
          </div>

          <div className="grid gap-3">
            {removedColumns.map((removedColumn) => {
              const removal = deletedColumns.find((item) => item.id === removedColumn.id);
              const targetOptions = availableReassignmentTargets(removedColumn.id);

              return (
                <div key={removedColumn.id} className="border-warning/30 bg-card/80 rounded-xl border p-3">
                  <div className="flex flex-col gap-3 lg:flex-row lg:items-end">
                    <div className="min-w-0 flex-1">
                      <p className="text-foreground text-sm font-medium">{removedColumn.name}</p>
                      <p className="text-muted-foreground text-xs">Tasks in this column need a destination.</p>
                    </div>

                    <div className="w-full lg:max-w-xs">
                      <Select
                        id={`reassign-${removedColumn.id}`}
                        label="Move tasks to"
                        value={removal?.move_tasks_to_column_id || targetOptions[0]?.value || ''}
                        onChange={(value) =>
                          onDeletedColumnsChange(
                            deletedColumns.map((item) =>
                              item.id === removedColumn.id ? { ...item, move_tasks_to_column_id: value } : item,
                            ),
                          )
                        }
                        options={targetOptions}
                        placeholder="Choose a destination"
                      />
                    </div>

                    <Button
                      type="button"
                      variant="secondary"
                      className="px-3 py-2"
                      onClick={() => restoreRemovedColumn(removedColumn)}
                    >
                      <RotateCcw className="h-4 w-4" />
                      Restore
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {error && <p className="text-destructive text-sm">{error}</p>}
    </section>
  );
};

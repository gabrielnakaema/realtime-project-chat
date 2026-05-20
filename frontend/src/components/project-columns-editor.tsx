import { ArrowDown, ArrowUp, CheckCircle2, Flag, Plus, RotateCcw, Trash2 } from 'lucide-react';
import { Button } from './button';
import { Input } from './input';
import { Select } from './select';
import type { IProjectForm } from '@/schemas/project-schema';
import type { ProjectColumn } from '@/types/project';
import { cn } from '@/lib/utils';
import { buildProjectColumnSurface, getDefaultProjectColumnColor } from '@/lib/project-column-colors';

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

const defaultColumns = (): FormColumn[] => [
  { name: 'Pending', color: getDefaultProjectColumnColor(0), is_done_column: false },
  { name: 'Doing', color: getDefaultProjectColumnColor(1), is_done_column: false },
  { name: 'Done', color: getDefaultProjectColumnColor(2), is_done_column: true },
];

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
    <section className="space-y-4 rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-slate-700 dark:bg-slate-900/50">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Flag className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Board columns</h3>
          </div>
          <p className="text-sm text-slate-600 dark:text-slate-400">
            Define the flow for this project. Every project needs at least one column and exactly one done column.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {mode === 'create' && (
            <Button
              type="button"
              variant="secondary"
              className="px-3 py-2 text-sm"
              onClick={() => onChange(defaultColumns())}
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
                { name: '', color: getDefaultProjectColumnColor(columns.length), is_done_column: false },
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

          return (
            <div
              key={getColumnKey(column, index)}
              style={{ borderColor: column.is_done_column ? undefined : surface.borderColor }}
              className={cn(
                'rounded-2xl border bg-white p-4 shadow-sm transition-colors dark:bg-slate-950',
                column.is_done_column
                  ? 'border-emerald-300/80 ring-1 ring-emerald-200 dark:border-emerald-700 dark:ring-emerald-900/60'
                  : 'border-slate-200 dark:border-slate-700',
              )}
            >
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start">
                <div className="flex min-w-0 flex-1 items-start gap-3">
                  <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-sm font-semibold text-slate-700 dark:bg-slate-800 dark:text-slate-200">
                    {index + 1}
                  </div>

                  <div className="min-w-0 flex-1">
                    <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_150px]">
                      <Input
                        id={`column-${index}`}
                        label={index === 0 ? 'Column name' : ' '}
                        placeholder={index === 0 ? 'Backlog' : 'Column name'}
                        value={column.name}
                        onChange={(event) => updateColumn(index, { name: event.target.value })}
                      />

                      <div className="space-y-2">
                        {index === 0 && (
                          <label
                            htmlFor={`column-color-${index}`}
                            className="block text-sm font-medium text-slate-700 dark:text-slate-300"
                          >
                            Color
                          </label>
                        )}
                        {index !== 0 && <div className="h-5" aria-hidden="true" />}
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
                            <p className="text-xs tracking-wide text-slate-500 uppercase dark:text-slate-400">
                              Main color
                            </p>
                            <p className="font-mono text-sm text-slate-800 dark:text-slate-100">{column.color}</p>
                          </div>
                        </div>
                      </div>
                    </div>

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
                            ? 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300'
                            : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:text-slate-900 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:text-slate-100',
                        )}
                      >
                        <CheckCircle2 className="h-4 w-4" />
                        {column.is_done_column ? 'Done column' : 'Mark as done'}
                      </button>

                      {column.id && (
                        <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600 dark:bg-slate-800 dark:text-slate-300">
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
                    onClick={() => moveColumn(index, -1)}
                    disabled={index === 0}
                  >
                    <ArrowUp className="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    className="px-3 py-2"
                    onClick={() => moveColumn(index, 1)}
                    disabled={index === columns.length - 1}
                  >
                    <ArrowDown className="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    className="px-3 py-2"
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
        <div className="space-y-3 rounded-2xl border border-amber-200 bg-amber-50/80 p-4 dark:border-amber-900 dark:bg-amber-950/40">
          <div className="space-y-1">
            <h4 className="text-sm font-semibold text-amber-900 dark:text-amber-100">Pending removals</h4>
            <p className="text-sm text-amber-800/80 dark:text-amber-200/80">
              Choose where tasks should move before these columns are removed.
            </p>
          </div>

          <div className="grid gap-3">
            {removedColumns.map((removedColumn) => {
              const removal = deletedColumns.find((item) => item.id === removedColumn.id);
              const targetOptions = availableReassignmentTargets(removedColumn.id!);

              return (
                <div
                  key={removedColumn.id}
                  className="rounded-xl border border-amber-200 bg-white/80 p-3 dark:border-amber-900 dark:bg-slate-950/70"
                >
                  <div className="flex flex-col gap-3 lg:flex-row lg:items-end">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{removedColumn.name}</p>
                      <p className="text-xs text-slate-600 dark:text-slate-400">
                        Tasks in this column need a destination.
                      </p>
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

      {error && <p className="text-sm text-red-500">{error}</p>}
    </section>
  );
};

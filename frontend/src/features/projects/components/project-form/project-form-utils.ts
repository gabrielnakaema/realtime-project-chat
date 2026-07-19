import type { IProjectForm } from '@/features/projects/schemas/project.schema';
import type { Project, ProjectColumn } from '@/features/projects/types/project';
import { getDefaultProjectColumnColor } from '@/features/projects/utils/project-column-colors';

type ProjectFormColumn = IProjectForm['columns'][number];
type ColumnKeyItem = Pick<ProjectColumn, 'id'> | Pick<ProjectFormColumn, 'id'>;

export const createDefaultProjectColumns = (): ProjectFormColumn[] => [
  { name: 'Pending', description: '', color: getDefaultProjectColumnColor(0), is_done_column: false },
  { name: 'Doing', description: '', color: getDefaultProjectColumnColor(1), is_done_column: false },
  { name: 'Done', description: '', color: getDefaultProjectColumnColor(2), is_done_column: true },
];

export const getProjectColumnKey = (column: ColumnKeyItem, index: number) => column.id ?? `new-column-${index}`;

export const restoreMissingProjectColumnIds = (
  columns: IProjectForm['columns'],
  originalColumns: Project['columns'],
  deletedColumns: IProjectForm['deleted_columns'] = [],
): IProjectForm['columns'] => {
  if (deletedColumns.length > 0 || columns.length !== originalColumns.length) {
    return columns;
  }

  const hasMissingIds = columns.some((column) => !column.id);
  if (!hasMissingIds) {
    return columns;
  }

  const hasAnyIds = columns.some((column) => column.id);
  if (hasAnyIds) {
    return columns;
  }

  return columns.map((column, index) => ({
    ...column,
    id: originalColumns[index]?.id,
  }));
};

export const getProjectFormValues = (project: Project): IProjectForm => ({
  name: project.name,
  description: project.description,
  repository_url: project.repository_url,
  repository_owner: project.repository_owner,
  repository_name: project.repository_name,
  default_branch: project.default_branch,
  branch_name_prefix: project.branch_name_prefix,
  columns: project.columns.map((column) => ({
    id: column.id,
    name: column.name,
    description: column.description,
    color: column.color,
    is_done_column: column.is_done_column,
  })),
  deleted_columns: [],
});

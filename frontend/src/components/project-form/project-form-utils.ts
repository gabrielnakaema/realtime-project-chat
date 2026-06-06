import type { IProjectForm } from '@/schemas/project-schema';
import type { Project, ProjectColumn } from '@/types/project';
import { getDefaultProjectColumnColor } from '@/lib/project-column-colors';

type ProjectFormColumn = IProjectForm['columns'][number];
type ColumnKeyItem = Pick<ProjectColumn, 'id'> | Pick<ProjectFormColumn, 'id'>;

export const createDefaultProjectColumns = (): ProjectFormColumn[] => [
  { name: 'Pending', description: '', color: getDefaultProjectColumnColor(0), is_done_column: false },
  { name: 'Doing', description: '', color: getDefaultProjectColumnColor(1), is_done_column: false },
  { name: 'Done', description: '', color: getDefaultProjectColumnColor(2), is_done_column: true },
];

export const getProjectColumnKey = (column: ColumnKeyItem, index: number) => column.id ?? `new-column-${index}`;

export const getProjectFormValues = (project: Project): IProjectForm => ({
  name: project.name,
  description: project.description,
  columns: project.columns.map((column) => ({
    id: column.id,
    name: column.name,
    description: column.description,
    color: column.color,
    is_done_column: column.is_done_column,
  })),
  deleted_columns: [],
});

import { z } from 'zod';

export interface IProjectForm {
  name: string;
  description: string;
  repository_url: string;
  repository_owner: string;
  repository_name: string;
  default_branch: string;
  branch_name_prefix: string;
  columns: {
    id?: string;
    name: string;
    description: string;
    color: string;
    is_done_column: boolean;
  }[];
  deleted_columns: {
    id: string;
    move_tasks_to_column_id: string;
  }[];
}

export const projectSchema: z.ZodType<IProjectForm> = z.object({
  name: z
    .string({
      error: 'Name is required',
    })
    .nonempty({ message: 'Name is required' }),
  description: z
    .string({
      error: 'Description is required',
    })
    .nonempty({ message: 'Description is required' }),
  repository_url: z.string(),
  repository_owner: z.string(),
  repository_name: z.string(),
  default_branch: z.string(),
  branch_name_prefix: z.string(),
  columns: z
    .array(
      z.object({
        id: z.string().optional(),
        name: z.string().nonempty({ message: 'Column name is required' }),
        description: z.string(),
        color: z.string().regex(/^#[0-9A-Fa-f]{6}$/, { message: 'Column color must be a valid hex value' }),
        is_done_column: z.boolean(),
      }),
    )
    .min(1, { message: 'At least one column is required' })
    .refine((columns) => columns.filter((column) => column.is_done_column).length === 1, {
      message: 'Exactly one done column is required',
    }),
  deleted_columns: z
    .array(
      z.object({
        id: z.string(),
        move_tasks_to_column_id: z.string(),
      }),
    )
    .default([]),
});

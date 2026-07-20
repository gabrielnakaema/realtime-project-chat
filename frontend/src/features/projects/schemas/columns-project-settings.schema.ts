import { z } from 'zod';

export type ColumnsProjectSettingsFormData = z.infer<typeof columnsProjectSettingsSchema>;

export const columnsProjectSettingsSchema = z.object({
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
  deleted_columns: z.array(
    z.object({
      id: z.string(),
      move_tasks_to_column_id: z.string(),
    }),
  ),
});

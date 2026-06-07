import { z } from 'zod';

export type IEditColumnForm = z.infer<typeof editColumnSchema>;

export const editColumnSchema = z.object({
  name: z.string().trim().min(1, 'Column name is required'),
  description: z.string(),
  color: z.string().regex(/^#[0-9A-Fa-f]{6}$/, { message: 'Column color must be a valid hex value' }),
  is_done_column: z.boolean(),
});

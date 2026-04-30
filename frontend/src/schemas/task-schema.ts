import { z } from 'zod';

export type ITaskForm = z.infer<typeof taskSchema>;

export const taskSchema = z.object({
  project_column_id: z.string().optional(),
  title: z
    .string({
      error: 'Title is required',
    })
    .nonempty({ message: 'Title is required' }),
  description: z
    .string({
      error: 'Description is required',
    })
    .nonempty({ message: 'Description is required' }),
  priority: z
    .string({
      error: 'Priority is required',
    })
    .nonempty({ message: 'Priority is required' }),
  responsible_id: z.string().optional().nullable(),
  due_date: z.string().optional().nullable(),
  tags: z.string().optional().nullable(),
});

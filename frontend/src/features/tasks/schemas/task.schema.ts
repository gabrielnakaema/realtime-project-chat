import { z } from 'zod';

export interface ITaskForm {
  project_column_id?: string;
  code?: string | null;
  title: string;
  description: string;
  priority: string;
  responsible_id?: string | null;
  due_date?: string | null;
  tags?: string | null;
  depends_on_task_ids: string[];
}

export const taskSchema: z.ZodType<ITaskForm> = z.object({
  project_column_id: z.string().optional(),
  code: z.string().trim().optional().nullable(),
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
  depends_on_task_ids: z.array(z.string()).default([]),
});

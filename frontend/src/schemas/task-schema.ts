import { z } from 'zod';

export type ITaskForm = z.infer<typeof taskSchema>;

export const taskSchema = z.object({
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
});

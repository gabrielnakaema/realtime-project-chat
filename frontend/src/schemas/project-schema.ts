import { z } from 'zod';

export type IProjectForm = z.infer<typeof projectSchema>;

export const projectSchema = z.object({
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
});

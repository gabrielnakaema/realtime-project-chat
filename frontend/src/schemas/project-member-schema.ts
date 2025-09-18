import { z } from 'zod';

export type IProjectMemberForm = z.infer<typeof projectMemberSchema>;

export const projectMemberSchema = z.object({
  email: z.email({
    error: 'Invalid email address',
  }),
});

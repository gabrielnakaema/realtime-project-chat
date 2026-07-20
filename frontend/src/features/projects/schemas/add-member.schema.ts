import { z } from 'zod';

export type AddMemberFormData = z.infer<typeof addMemberSchema>;

export const addMemberSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Email is required.')
    .pipe(z.email({ error: 'Enter a valid email address.' })),
});

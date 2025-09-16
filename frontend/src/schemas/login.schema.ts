import { z } from 'zod';

export type ILoginForm = z.infer<typeof loginSchema>;

export const loginSchema = z.object({
  email: z
    .email({
      error: 'Invalid email address',
    })
    .nonempty({
      error: 'Email is required',
    }),
  password: z
    .string({
      error: 'Password is required',
    })
    .nonempty({
      error: 'Password is required',
    }),
});

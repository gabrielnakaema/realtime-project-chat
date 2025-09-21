import { z } from 'zod';

export type ISignUpForm = z.infer<typeof signUpSchema>;

export const signUpSchema = z
  .object({
    name: z
      .string({
        error: 'Name is required',
      })
      .nonempty({
        error: 'Name is required',
      }),
    email: z.email({
      error: 'Invalid email address',
    }),
    password: z
      .string({
        error: 'Password is required',
      })
      .nonempty({
        error: 'Password is required',
      })
      .min(8, {
        message: 'Password must be at least 8 characters',
      }),
    confirmPassword: z
      .string({
        error: 'Confirm your password',
      })
      .nonempty({
        error: 'Confirm your password',
      }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Passwords do not match',
  });

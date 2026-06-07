import { z } from 'zod';

export type IChangePasswordForm = z.infer<typeof changePasswordSchema>;

export const changePasswordSchema = z
  .object({
    old_password: z
      .string({
        error: 'old password is required',
      })
      .nonempty({
        error: 'old password is required',
      }),
    new_password: z
      .string({
        error: 'new password is required',
      })
      .nonempty({
        error: 'new password is required',
      })
      .min(6, {
        message: 'password must be at least 6 characters',
      }),
    new_password_confirmation: z
      .string({
        error: 'new password confirmation is required',
      })
      .nonempty({
        error: 'new password confirmation is required',
      }),
  })
  .refine((data) => data.new_password === data.new_password_confirmation, {
    path: ['new_password_confirmation'],
    message: 'new password confirmation must match new password',
  });

import { z } from 'zod';

export type IChatForm = z.infer<typeof chatSchema>;

export const chatSchema = z.object({
  content: z
    .string({
      error: 'Content is required',
    })
    .nonempty({
      message: 'Content is required',
    }),
});

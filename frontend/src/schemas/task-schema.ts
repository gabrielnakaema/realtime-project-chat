import { endOfDay, isAfter, parse, startOfDay } from 'date-fns';
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
  priority: z
    .string({
      error: 'Priority is required',
    })
    .nonempty({ message: 'Priority is required' }),
  responsible_id: z.string().optional().nullable(),
  due_date: z
    .string()
    .optional()
    .nullable()
    .refine(
      (date) => {
        if (!date) return true;
        try {
          const parsedDate = parse(date, 'yyyy-MM-dd', new Date());
          return isAfter(endOfDay(parsedDate), startOfDay(new Date()));
        } catch (err) {
          return false;
        }
      },
      { message: 'Due date cannot be in the past' },
    ),
  tags: z.string().optional().nullable(),
});

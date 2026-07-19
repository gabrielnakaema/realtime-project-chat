import { z } from 'zod';

export const mcpAPIKeySchema = z.object({
  name: z
    .string({
      error: 'Name is required',
    })
    .trim()
    .min(1, { message: 'Name is required' }),
  scopes: z.array(z.string().trim().min(1)).min(1, { message: 'Select at least one scope' }),
});

export type IMCPAPIKeyForm = z.infer<typeof mcpAPIKeySchema>;

export const mcpAPIKeyDefaultValues: IMCPAPIKeyForm = {
  name: '',
  scopes: [],
};

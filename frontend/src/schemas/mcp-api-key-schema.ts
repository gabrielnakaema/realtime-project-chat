import { z } from 'zod';
import { mcpApiScopeValues } from '@/services/mcp-api-keys';

export const createMCPAPIKeySchema = z.object({
  name: z
    .string({
      error: 'Name is required',
    })
    .trim()
    .min(1, { message: 'Name is required' }),
  scopes: z.array(z.enum(mcpApiScopeValues)).min(1, { message: 'Select at least one scope' }),
});

export type ICreateMCPAPIKeyForm = z.infer<typeof createMCPAPIKeySchema>;

export const createMCPAPIKeyDefaultValues: ICreateMCPAPIKeyForm = {
  name: '',
  scopes: [],
};

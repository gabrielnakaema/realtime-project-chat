import { z } from 'zod';

export type GeneralProjectSettingsFormData = z.infer<typeof generalProjectSettingsSchema>;

export const generalProjectSettingsSchema = z.object({
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
  repository_url: z.string(),
  repository_owner: z.string(),
  repository_name: z.string(),
  default_branch: z.string(),
  branch_name_prefix: z.string(),
});

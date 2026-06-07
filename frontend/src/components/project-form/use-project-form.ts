import { zodResolver } from '@hookform/resolvers/zod';
import { useCallback } from 'react';
import { useForm } from 'react-hook-form';
import type { IProjectForm } from '@/schemas/project-schema';
import { createDefaultProjectColumns } from '@/components/project-form/project-form-utils';
import { projectSchema } from '@/schemas/project-schema';

const getBaseProjectFormValues = (): IProjectForm => ({
  name: '',
  description: '',
  repository_url: '',
  repository_owner: '',
  repository_name: '',
  default_branch: '',
  branch_name_prefix: '',
  columns: createDefaultProjectColumns(),
  deleted_columns: [],
});

export const useProjectForm = () => {
  const form = useForm<IProjectForm>({
    resolver: zodResolver(projectSchema as any) as any,
    defaultValues: getBaseProjectFormValues(),
  });
  const { reset } = form;

  const resetForm = useCallback(
    (values?: Partial<IProjectForm>) => {
      reset({
        ...getBaseProjectFormValues(),
        ...values,
      });
    },
    [reset],
  );

  return {
    form,
    resetForm,
  };
};

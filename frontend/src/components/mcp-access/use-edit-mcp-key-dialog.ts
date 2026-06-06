import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import { sortMCPAPIKeys } from './mcp-access-utils';
import type { QueryClient } from '@tanstack/react-query';
import type { IMCPAPIKeyForm } from '@/schemas/mcp-api-key-schema';
import type { MCPAPIKey } from '@/services/mcp-api-keys';
import { updateMCPAPIKey } from '@/services/mcp-api-keys';
import { mcpAPIKeyQueryKeys } from '@/services/query-keys';
import { getErrorMessage } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';

interface UseEditMcpKeyDialogOptions {
  keyData: MCPAPIKey;
  onOpenChange: (open: boolean) => void;
}

export const useEditMcpKeyDialog = ({ keyData, onOpenChange }: UseEditMcpKeyDialogOptions) => {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const defaultValues = useMemo<IMCPAPIKeyForm>(
    () => ({
      name: keyData.name,
      scopes: keyData.scopes,
    }),
    [keyData.name, keyData.scopes],
  );

  const handleDialogOpenChange = (open: boolean) => {
    if (!open) {
      setErrorMessage(null);
    }

    onOpenChange(open);
  };

  const updateMutation = useMutation({
    mutationFn: (form: IMCPAPIKeyForm) => updateMCPAPIKey(keyData.id, form),
    onSuccess: (updatedKey) => {
      updateCachedKey(queryClient, updatedKey);
      handleSuccess('Key updated');
      onOpenChange(false);
    },
    onError: async (error) => {
      setErrorMessage(await getErrorMessage(error));
    },
  });

  return {
    defaultValues,
    errorMessage,
    handleDialogOpenChange,
    isSubmitting: updateMutation.isPending,
    resetErrorMessage: () => setErrorMessage(null),
    submitUpdate: updateMutation.mutate,
  };
};

const updateCachedKey = (queryClient: QueryClient, updatedKey: MCPAPIKey) => {
  queryClient.setQueryData<MCPAPIKey[]>(mcpAPIKeyQueryKeys.list, (current) => {
    if (!current) {
      return [updatedKey];
    }

    const next = current.map((key) => (key.id === updatedKey.id ? updatedKey : key));
    return sortMCPAPIKeys(next);
  });

  queryClient.invalidateQueries({ queryKey: mcpAPIKeyQueryKeys.all });
};

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import { sortMCPAPIKeys } from '../components/mcp-access-utils';
import type { QueryClient } from '@tanstack/react-query';
import type { IMCPAPIKeyForm } from '@/features/mcp-access/schemas/mcp-api-key.schema';
import type { MCPAPIKey } from '@/features/mcp-access/services/mcp-api-keys';
import { updateMCPAPIKey } from '@/features/mcp-access/services/mcp-api-keys';
import { mcpAPIKeyQueryKeys } from '@/shared/services/query-keys';
import { getErrorMessage } from '@/shared/utils/handle-error';
import { handleSuccess } from '@/shared/utils/handle-success';

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

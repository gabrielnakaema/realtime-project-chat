import { useMutation, useQueryClient } from '@tanstack/react-query';
import { KeyRound } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '../ui/dialog';
import { MCPAPIKeyForm } from './mcp-key-form';
import { sortMCPAPIKeys } from './mcp-access-utils';
import type { IMCPAPIKeyForm } from '@/schemas/mcp-api-key-schema';
import type { MCPAPIKey } from '@/services/mcp-api-keys';
import { updateMCPAPIKey } from '@/services/mcp-api-keys';
import { mcpAPIKeyQueryKeys } from '@/services/query-keys';
import { getErrorMessage } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';

interface EditMCPKeyDialogProps {
  isOpen: boolean;
  keyData: MCPAPIKey;
  onOpenChange: (open: boolean) => void;
}

export const EditMCPKeyDialog = ({ isOpen, keyData, onOpenChange }: EditMCPKeyDialogProps) => {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const defaultValues = useMemo<IMCPAPIKeyForm>(
    () => ({
      name: keyData.name,
      scopes: keyData.scopes,
    }),
    [keyData.name, keyData.scopes],
  );

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
  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setErrorMessage(null);
    }

    onOpenChange(open);
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col gap-0 overflow-hidden p-0 sm:max-w-3xl">
        <DialogHeader className="border-b border-slate-200 bg-slate-50/80 px-6 py-5 dark:border-slate-800 dark:bg-slate-900/80">
          <DialogTitle className="flex items-center gap-2">
            <KeyRound className="h-5 w-5 text-blue-600" />
            Edit MCP key
          </DialogTitle>
          <DialogDescription>Update the key name or permissions without rotating the secret.</DialogDescription>
        </DialogHeader>

        <MCPAPIKeyForm
          defaultValues={defaultValues}
          errorMessage={errorMessage}
          isSubmitting={updateMutation.isPending}
          submitLabel="Save changes"
          onCancel={() => handleOpenChange(false)}
          onFormChange={() => setErrorMessage(null)}
          onSubmit={updateMutation.mutate}
        />
      </DialogContent>
    </Dialog>
  );
};

const updateCachedKey = (queryClient: ReturnType<typeof useQueryClient>, updatedKey: MCPAPIKey) => {
  queryClient.setQueryData<MCPAPIKey[]>(mcpAPIKeyQueryKeys.list, (current) => {
    if (!current) {
      return [updatedKey];
    }

    const next = current.map((key) => (key.id === updatedKey.id ? updatedKey : key));
    return sortMCPAPIKeys(next);
  });

  queryClient.invalidateQueries({ queryKey: mcpAPIKeyQueryKeys.all });
};

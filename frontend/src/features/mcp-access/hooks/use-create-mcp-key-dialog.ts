import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { sortMCPAPIKeys } from '../components/mcp-access-utils';
import type { CreateMCPAPIKeyResponse, MCPAPIKey } from '@/features/mcp-access/services/mcp-api-keys';
import type { QueryClient } from '@tanstack/react-query';
import type { CreateModalMode } from '../components/mcp-access-utils';
import { createMCPAPIKey, getMCPServerURL } from '@/features/mcp-access/services/mcp-api-keys';
import { useCopyFeedback } from '@/shared/hooks/use-copy-feedback';
import { mcpAPIKeyQueryKeys } from '@/shared/services/query-keys';
import { copyToClipboard } from '@/shared/utils/clipboard';
import { getErrorMessage } from '@/shared/utils/handle-error';
import { handleSuccess } from '@/shared/utils/handle-success';

export type CopyType = 'secret' | 'url';

export const useCreateMcpKeyDialog = () => {
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState<CreateModalMode>('create');
  const [rawSecret, setRawSecret] = useState('');
  const [hasCopiedSecret, setHasCopiedSecret] = useState(false);
  const [dismissRevealPromptOpen, setDismissRevealPromptOpen] = useState(false);
  const [createErrorMessage, setCreateErrorMessage] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const serverUrl = getMCPServerURL();
  const { copiedValue, markCopied, resetCopiedValue } = useCopyFeedback<CopyType>();

  const isRevealMode = mode === 'reveal';

  const resetDialogState = () => {
    setMode('create');
    setRawSecret('');
    setHasCopiedSecret(false);
    setDismissRevealPromptOpen(false);
    setCreateErrorMessage(null);
    resetCopiedValue();
  };

  const closeDialog = () => {
    resetDialogState();
    setOpen(false);
  };

  const shouldPromptBeforeClose = isRevealMode && rawSecret !== '' && !hasCopiedSecret;

  const handleDialogOpenChange = (nextOpen: boolean) => {
    if (nextOpen) {
      setOpen(true);
      return;
    }

    if (shouldPromptBeforeClose) {
      setDismissRevealPromptOpen(true);
      return;
    }

    closeDialog();
  };

  const copyValue = async (value: string, type: CopyType) => {
    await copyToClipboard(value);
    markCopied(type);

    if (type === 'secret') {
      setHasCopiedSecret(true);
      handleSuccess('Secret copied');
      return;
    }

    handleSuccess('Server URL copied');
  };

  const createMutation = useMutation({
    mutationFn: createMCPAPIKey,
    onSuccess: (response: CreateMCPAPIKeyResponse) => {
      setRawSecret(response.raw_secret);
      setHasCopiedSecret(false);
      resetCopiedValue();
      setMode('reveal');
      updateCachedKeys(queryClient, response);
    },
    onError: async (error) => {
      setCreateErrorMessage(await getErrorMessage(error));
    },
  });

  return {
    closeDialog,
    copiedValue,
    copyValue,
    createErrorMessage,
    createKey: createMutation.mutate,
    dismissRevealPromptOpen,
    handleDialogOpenChange,
    isRevealMode,
    isSubmitting: createMutation.isPending,
    keepRevealOpen: () => setDismissRevealPromptOpen(false),
    open,
    rawSecret,
    resetErrorMessage: () => setCreateErrorMessage(null),
    serverUrl,
  };
};

const updateCachedKeys = (queryClient: QueryClient, response: CreateMCPAPIKeyResponse) => {
  queryClient.setQueryData<MCPAPIKey[]>(mcpAPIKeyQueryKeys.list, (current) => {
    const next = current ? current.filter((key) => key.id !== response.key.id) : [];
    return sortMCPAPIKeys([response.key, ...next]);
  });

  queryClient.invalidateQueries({ queryKey: mcpAPIKeyQueryKeys.all });
};

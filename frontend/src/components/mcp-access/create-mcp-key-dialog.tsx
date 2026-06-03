import { useMutation, useQueryClient } from '@tanstack/react-query';
import { AlertTriangle, Check, Copy, KeyRound, Plus } from 'lucide-react';
import { useState } from 'react';
import { Button } from '../button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../ui/dialog';
import { MCPAPIKeyForm } from './mcp-key-form';
import { sortMCPAPIKeys } from './mcp-access-utils';
import { MCPHelpDialog } from './mcp-help-dialog';
import type { CreateModalMode } from './mcp-access-utils';
import type { CreateMCPAPIKeyResponse, MCPAPIKey } from '@/services/mcp-api-keys';
import { mcpAPIKeyDefaultValues } from '@/schemas/mcp-api-key-schema';
import { createMCPAPIKey, getMCPServerURL } from '@/services/mcp-api-keys';
import { mcpAPIKeyQueryKeys } from '@/services/query-keys';
import { useCopyFeedback } from '@/hooks/use-copy-feedback';
import { getErrorMessage } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';
import { copyToClipboard } from '@/utils/clipboard';

interface CreateMCPKeyButtonProps {
  className?: string;
}

export const CreateMCPKeyButton = ({ className }: CreateMCPKeyButtonProps) => {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const serverUrl = getMCPServerURL();
  const { copiedValue, markCopied, resetCopiedValue } = useCopyFeedback<'secret' | 'url'>();
  const [mode, setMode] = useState<CreateModalMode>('create');
  const [rawSecret, setRawSecret] = useState('');
  const [hasCopiedSecret, setHasCopiedSecret] = useState(false);
  const [dismissRevealPromptOpen, setDismissRevealPromptOpen] = useState(false);
  const [createErrorMessage, setCreateErrorMessage] = useState<string | null>(null);

  const isRevealMode = mode === 'reveal';

  const clearCreateState = () => {
    setMode('create');
    setRawSecret('');
    setHasCopiedSecret(false);
    setDismissRevealPromptOpen(false);
    setCreateErrorMessage(null);
    resetCopiedValue();
  };

  const closeDialog = () => {
    clearCreateState();
    setOpen(false);
  };

  const handleDialogOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && mode === 'reveal' && rawSecret && !hasCopiedSecret) {
      setDismissRevealPromptOpen(true);
      return;
    }

    if (!nextOpen) {
      closeDialog();
      return;
    }

    setOpen(true);
  };

  const copyText = async (value: string, type: 'secret' | 'url') => {
    await copyToClipboard(value);
    markCopied(type);

    if (type === 'secret') {
      setHasCopiedSecret(true);
      handleSuccess('Secret copied');
      return;
    }

    handleSuccess('Server URL copied');
  };

  const handleCreated = (response: CreateMCPAPIKeyResponse) => {
    setRawSecret(response.raw_secret);
    setHasCopiedSecret(false);
    resetCopiedValue();
    setMode('reveal');
    updateCachedKeys(queryClient, response);
  };

  const createMutation = useMutation({
    mutationFn: createMCPAPIKey,
    onSuccess: handleCreated,
    onError: async (error) => {
      setCreateErrorMessage(await getErrorMessage(error));
    },
  });

  return (
    <>
      <Dialog open={open} onOpenChange={handleDialogOpenChange}>
        <DialogTrigger asChild>
          <Button type="button" className={className}>
            <Plus className="h-4 w-4" />
            Create key
          </Button>
        </DialogTrigger>
        <DialogContent
          className="flex max-h-[calc(100dvh-2rem)] flex-col gap-0 overflow-hidden p-0 sm:max-w-3xl"
          onEscapeKeyDown={isRevealMode ? (event) => event.preventDefault() : undefined}
          onPointerDownOutside={isRevealMode ? (event) => event.preventDefault() : undefined}
        >
          <DialogHeader className="border-b border-slate-200 bg-slate-50/80 px-6 py-5 dark:border-slate-800 dark:bg-slate-900/80">
            <DialogTitle className="flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-blue-600" />
              {isRevealMode ? 'Save your new MCP key' : 'Create MCP key'}
            </DialogTitle>
            <DialogDescription>
              {isRevealMode
                ? 'Copy the secret now. It will not be shown again after you close this modal.'
                : 'Name the key and choose the permissions this MCP client should receive.'}
            </DialogDescription>
          </DialogHeader>

          {isRevealMode ? (
            <>
              <div className="flex-1 overflow-y-auto px-6 py-5">
                <RevealSecretContent
                  copiedValue={copiedValue}
                  onCopy={copyText}
                  rawSecret={rawSecret}
                  serverUrl={serverUrl}
                />
              </div>

              <DialogFooter className="border-t border-slate-200 bg-slate-50/80 px-6 py-4 dark:border-slate-800 dark:bg-slate-900/80">
                <Button type="button" onClick={closeDialog}>
                  Close
                </Button>
              </DialogFooter>
            </>
          ) : (
            <MCPAPIKeyForm
              defaultValues={mcpAPIKeyDefaultValues}
              errorMessage={createErrorMessage}
              isSubmitting={createMutation.isPending}
              submitLabel="Create key"
              onCancel={closeDialog}
              onFormChange={() => setCreateErrorMessage(null)}
              onSubmit={createMutation.mutate}
            />
          )}
        </DialogContent>
      </Dialog>

      <Dialog open={dismissRevealPromptOpen} onOpenChange={setDismissRevealPromptOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Leave without copying?</DialogTitle>
            <DialogDescription>
              This secret will not be shown again. If you close now, you will need to create a new key to reconnect the
              client.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="secondary" onClick={() => setDismissRevealPromptOpen(false)}>
              Keep this open
            </Button>
            <Button type="button" onClick={closeDialog}>
              Close anyway
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};

const RevealSecretContent = ({
  copiedValue,
  onCopy,
  rawSecret,
  serverUrl,
}: {
  copiedValue: 'secret' | 'url' | null;
  onCopy: (value: string, type: 'secret' | 'url') => Promise<void>;
  rawSecret: string;
  serverUrl: string;
}) => {
  return (
    <section className="space-y-5">
      <div className="rounded-2xl border border-amber-300 bg-amber-50 p-4 text-amber-950 dark:border-amber-700/70 dark:bg-amber-950/40 dark:text-amber-100">
        <div className="flex items-start gap-3">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0" />
          <div className="space-y-1">
            <p className="font-semibold">Shown once</p>
            <p className="text-sm">
              Copy and store this key now. For security, it will not be shown again after you leave this step.
            </p>
          </div>
        </div>
      </div>

      <div className="flex justify-end">
        <MCPHelpDialog serverUrl={serverUrl} rawSecret={rawSecret} />
      </div>

      <CopyableSecretCard
        copiedValue={copiedValue}
        label="API key"
        helperText="Use this secret in your MCP client."
        value={rawSecret}
        copyType="secret"
        copyLabel="Copy key"
        copiedLabel="Copied"
        onCopy={onCopy}
      />

      <CopyableSecretCard
        copiedValue={copiedValue}
        label="Server URL"
        helperText="Point your MCP client to this endpoint."
        value={serverUrl}
        copyType="url"
        copyLabel="Copy URL"
        copiedLabel="Copied"
        onCopy={onCopy}
      />
    </section>
  );
};

const CopyableSecretCard = ({
  copiedLabel,
  copiedValue,
  copyLabel,
  copyType,
  helperText,
  label,
  onCopy,
  value,
}: {
  copiedLabel: string;
  copiedValue: 'secret' | 'url' | null;
  copyLabel: string;
  copyType: 'secret' | 'url';
  helperText: string;
  label: string;
  onCopy: (value: string, type: 'secret' | 'url') => Promise<void>;
  value: string;
}) => {
  const isCopied = copiedValue === copyType;

  return (
    <div className="space-y-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-950">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">{label}</p>
          <p className="text-sm text-slate-600 dark:text-slate-400">{helperText}</p>
        </div>
        <Button type="button" variant="secondary" className="shrink-0" onClick={() => void onCopy(value, copyType)}>
          {isCopied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          {isCopied ? copiedLabel : copyLabel}
        </Button>
      </div>
      <div className="overflow-x-auto rounded-xl border border-slate-200 bg-slate-50 p-3 font-mono text-sm break-all dark:border-slate-700 dark:bg-slate-900">
        {value}
      </div>
    </div>
  );
};

const updateCachedKeys = (queryClient: ReturnType<typeof useQueryClient>, response: CreateMCPAPIKeyResponse) => {
  queryClient.setQueryData<MCPAPIKey[]>(mcpAPIKeyQueryKeys.list, (current) => {
    const next = current ? current.filter((key) => key.id !== response.key.id) : [];
    return sortMCPAPIKeys([response.key, ...next]);
  });

  queryClient.invalidateQueries({ queryKey: mcpAPIKeyQueryKeys.all });
};

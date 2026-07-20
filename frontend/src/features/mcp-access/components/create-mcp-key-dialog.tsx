import { KeyRound, Plus } from 'lucide-react';
import { Button } from '../../../components/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../../../shared/components/ui/dialog';
import { useCreateMcpKeyDialog } from '../hooks/use-create-mcp-key-dialog';
import { CreateMCPKeyRevealStep, DismissRevealPromptDialog } from './create-mcp-key-dialog-steps';
import { MCPAPIKeyForm } from './mcp-key-form';
import { mcpAPIKeyDefaultValues } from '@/features/mcp-access/schemas/mcp-api-key.schema';

interface CreateMCPKeyButtonProps {
  className?: string;
}

export const CreateMCPKeyButton = ({ className }: CreateMCPKeyButtonProps) => {
  const {
    closeDialog,
    copiedValue,
    copyValue,
    createErrorMessage,
    createKey,
    dismissRevealPromptOpen,
    handleDialogOpenChange,
    isRevealMode,
    isSubmitting,
    keepRevealOpen,
    open,
    rawSecret,
    resetErrorMessage,
    serverUrl,
  } = useCreateMcpKeyDialog();

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
          <DialogHeader className="border-border bg-muted/80 border-b px-6 py-5">
            <DialogTitle className="flex items-center gap-2">
              <KeyRound className="text-primary h-5 w-5" />
              {isRevealMode ? 'Save your new MCP key' : 'Create MCP key'}
            </DialogTitle>
            <DialogDescription>
              {isRevealMode
                ? 'Copy the secret now. It will not be shown again after you close this modal.'
                : 'Name the key and choose the permissions this MCP client should receive.'}
            </DialogDescription>
          </DialogHeader>

          {isRevealMode ? (
            <CreateMCPKeyRevealStep
              copiedValue={copiedValue}
              onClose={() => handleDialogOpenChange(false)}
              onCopy={copyValue}
              rawSecret={rawSecret}
              serverUrl={serverUrl}
            />
          ) : (
            <MCPAPIKeyForm
              defaultValues={mcpAPIKeyDefaultValues}
              errorMessage={createErrorMessage}
              isSubmitting={isSubmitting}
              submitLabel="Create key"
              onCancel={closeDialog}
              onFormChange={resetErrorMessage}
              onSubmit={createKey}
            />
          )}
        </DialogContent>
      </Dialog>

      <DismissRevealPromptDialog
        open={dismissRevealPromptOpen}
        onCloseAnyway={closeDialog}
        onKeepOpen={keepRevealOpen}
      />
    </>
  );
};

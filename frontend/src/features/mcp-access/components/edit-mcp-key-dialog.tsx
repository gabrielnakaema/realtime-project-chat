import { KeyRound } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '../../../shared/components/ui/dialog';
import { useEditMcpKeyDialog } from '../hooks/use-edit-mcp-key-dialog';
import { MCPAPIKeyForm } from './mcp-key-form';
import type { MCPAPIKey } from '@/features/mcp-access/services/mcp-api-keys';

interface EditMCPKeyDialogProps {
  isOpen: boolean;
  keyData: MCPAPIKey;
  onOpenChange: (open: boolean) => void;
}

export const EditMCPKeyDialog = ({ isOpen, keyData, onOpenChange }: EditMCPKeyDialogProps) => {
  const { defaultValues, errorMessage, handleDialogOpenChange, isSubmitting, resetErrorMessage, submitUpdate } =
    useEditMcpKeyDialog({
      keyData,
      onOpenChange,
    });

  return (
    <Dialog open={isOpen} onOpenChange={handleDialogOpenChange}>
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col gap-0 overflow-hidden p-0 sm:max-w-3xl">
        <DialogHeader className="border-border bg-muted/80 border-b px-6 py-5">
          <DialogTitle className="flex items-center gap-2">
            <KeyRound className="text-primary h-5 w-5" />
            Edit MCP key
          </DialogTitle>
          <DialogDescription>Update the key name or permissions without rotating the secret.</DialogDescription>
        </DialogHeader>

        <MCPAPIKeyForm
          defaultValues={defaultValues}
          errorMessage={errorMessage}
          isSubmitting={isSubmitting}
          submitLabel="Save changes"
          onCancel={() => handleDialogOpenChange(false)}
          onFormChange={resetErrorMessage}
          onSubmit={submitUpdate}
        />
      </DialogContent>
    </Dialog>
  );
};

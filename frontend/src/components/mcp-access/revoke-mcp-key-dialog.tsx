import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog';
import type { MCPAPIKey } from '@/services/mcp-api-keys';
import { revokeMCPAPIKey } from '@/services/mcp-api-keys';
import { mcpAPIKeyQueryKeys } from '@/services/query-keys';
import { handleError } from '@/utils/handle-error';
import { handleSuccess } from '@/utils/handle-success';

interface RevokeMCPKeyDialogProps {
  isOpen: boolean;
  keyData: MCPAPIKey;
  onOpenChange: (open: boolean) => void;
}

export const RevokeMCPKeyDialog = ({ isOpen, keyData, onOpenChange }: RevokeMCPKeyDialogProps) => {
  const queryClient = useQueryClient();
  const revokeMutation = useMutation({
    mutationFn: revokeMCPAPIKey,
    onSuccess: (_, id) => {
      onOpenChange(false);
      queryClient.setQueryData<MCPAPIKey[]>(mcpAPIKeyQueryKeys.list, (current) =>
        (current ?? []).map((key) =>
          key.id === id
            ? {
                ...key,
                revoked_at: new Date().toISOString(),
              }
            : key,
        ),
      );
      queryClient.invalidateQueries({ queryKey: mcpAPIKeyQueryKeys.all });
      handleSuccess('MCP key revoked');
    },
    onError: (error) => {
      handleError(error);
      queryClient.invalidateQueries({ queryKey: mcpAPIKeyQueryKeys.all });
    },
  });

  const confirmRevoke = () => {
    if (revokeMutation.isPending) {
      return;
    }

    revokeMutation.mutate(keyData.id);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke MCP key</DialogTitle>
          <DialogDescription>
            Revoke <span className="text-foreground font-medium">{keyData.name}</span>? Existing MCP clients using this
            key will stop working immediately.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="secondary" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="button" disabled={revokeMutation.isPending} onClick={confirmRevoke}>
            {revokeMutation.isPending ? <LoadingSpinner size="1.25rem" /> : 'Revoke key'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

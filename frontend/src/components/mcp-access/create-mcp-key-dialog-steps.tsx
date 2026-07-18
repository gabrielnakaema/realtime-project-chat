import { AlertTriangle, Check, Copy } from 'lucide-react';
import { Button } from '../button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog';
import { MCPHelpDialog } from './mcp-help-dialog';
import type { CopyType } from './use-create-mcp-key-dialog';

interface CreateMCPKeyRevealStepProps {
  copiedValue: CopyType | null;
  onClose: () => void;
  onCopy: (value: string, type: CopyType) => Promise<void>;
  rawSecret: string;
  serverUrl: string;
}

export const CreateMCPKeyRevealStep = ({
  copiedValue,
  onClose,
  onCopy,
  rawSecret,
  serverUrl,
}: CreateMCPKeyRevealStepProps) => {
  return (
    <>
      <div className="flex-1 overflow-y-auto px-6 py-5">
        <section className="space-y-5">
          <div className="border-warning/30 bg-warning/10 text-warning rounded-2xl border p-4">
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
            copyLabel="Copy key"
            copyType="secret"
            helperText="Use this secret in your MCP client."
            label="API key"
            onCopy={onCopy}
            value={rawSecret}
          />

          <CopyableSecretCard
            copiedValue={copiedValue}
            copyLabel="Copy URL"
            copyType="url"
            helperText="Point your MCP client to this endpoint."
            label="Server URL"
            onCopy={onCopy}
            value={serverUrl}
          />
        </section>
      </div>

      <DialogFooter className="border-border bg-muted/80 border-t px-6 py-4">
        <Button type="button" onClick={onClose}>
          Close
        </Button>
      </DialogFooter>
    </>
  );
};

interface DismissRevealPromptDialogProps {
  onCloseAnyway: () => void;
  onKeepOpen: () => void;
  open: boolean;
}

export const DismissRevealPromptDialog = ({ onCloseAnyway, onKeepOpen, open }: DismissRevealPromptDialogProps) => {
  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) onKeepOpen();
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Leave without copying?</DialogTitle>
          <DialogDescription>
            This secret will not be shown again. If you close now, you will need to create a new key to reconnect the
            client.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="secondary" onClick={onKeepOpen}>
            Keep this open
          </Button>
          <Button type="button" onClick={onCloseAnyway}>
            Close anyway
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

interface CopyableSecretCardProps {
  copiedValue: CopyType | null;
  copyLabel: string;
  copyType: CopyType;
  helperText: string;
  label: string;
  onCopy: (value: string, type: CopyType) => Promise<void>;
  value: string;
}

const CopyableSecretCard = ({
  copiedValue,
  copyLabel,
  copyType,
  helperText,
  label,
  onCopy,
  value,
}: CopyableSecretCardProps) => {
  const isCopied = copiedValue === copyType;

  return (
    <div className="border-border bg-card space-y-3 rounded-2xl border p-4 shadow-sm">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-foreground text-sm font-semibold">{label}</p>
          <p className="text-muted-foreground text-sm">{helperText}</p>
        </div>
        <Button type="button" variant="secondary" className="shrink-0" onClick={() => void onCopy(value, copyType)}>
          {isCopied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          {isCopied ? 'Copied' : copyLabel}
        </Button>
      </div>
      <div className="border-border bg-muted overflow-x-auto rounded-xl border p-3 font-mono text-sm break-all">
        {value}
      </div>
    </div>
  );
};

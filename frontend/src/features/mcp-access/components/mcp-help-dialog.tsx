import { CircleQuestionMark } from 'lucide-react';
import { useState } from 'react';
import { Button } from '../../../components/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '../../../shared/components/ui/dialog';

const helpCodeBlockClassName = 'overflow-x-auto rounded-xl border border-border bg-muted p-3 font-mono text-sm  ';

export const MCPHelpDialog = ({ serverUrl, rawSecret }: { serverUrl: string; rawSecret?: string }) => {
  const [open, setOpen] = useState(false);

  const authorizationExample = rawSecret ? `Bearer ${rawSecret}` : 'Bearer <your_mcp_api_key>';

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Button type="button" variant="secondary" onClick={() => setOpen(true)}>
        <CircleQuestionMark className="h-4 w-4" />
        Help
      </Button>

      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>How to connect an MCP client</DialogTitle>
          <DialogDescription>
            Use the server URL and your API key together. The key should be sent as a Bearer token in the
            `Authorization` header.
          </DialogDescription>
        </DialogHeader>

        <div className="border-border bg-muted/80 space-y-4 rounded-2xl border p-4">
          <div className="space-y-1">
            <h3 className="text-foreground text-sm font-semibold">Quick setup help</h3>
            <p className="text-muted-foreground text-sm">
              Most MCP clients need the server URL plus your key in the `Authorization` header.
            </p>
          </div>

          <div className="space-y-2">
            <p className="text-foreground text-sm font-medium">Server URL</p>
            <div className={helpCodeBlockClassName}>{serverUrl}</div>
          </div>

          <div className="space-y-2">
            <p className="text-foreground text-sm font-medium">Authorization header</p>
            <div className={helpCodeBlockClassName}>Authorization: {authorizationExample}</div>
          </div>

          <div className="text-muted-foreground space-y-2 text-sm">
            <p>1. Copy the server URL into your MCP client configuration.</p>
            <p>2. Paste the generated key as a Bearer token in the `Authorization` header.</p>
            <p>3. Keep the key in a secure secret manager or client config, not in shared notes.</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

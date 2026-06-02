import { CircleQuestionMark } from 'lucide-react';
import { useState } from 'react';
import { Button } from '../button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '../ui/dialog';

const helpCodeBlockClassName =
  'overflow-x-auto rounded-xl border border-slate-200 bg-slate-50 p-3 font-mono text-sm dark:border-slate-700 dark:bg-slate-900';

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

        <div className="space-y-4 rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-slate-800 dark:bg-slate-900/60">
          <div className="space-y-1">
            <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Quick setup help</h3>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Most MCP clients need the server URL plus your key in the `Authorization` header.
            </p>
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium text-slate-900 dark:text-slate-100">Server URL</p>
            <div className={helpCodeBlockClassName}>{serverUrl}</div>
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium text-slate-900 dark:text-slate-100">Authorization header</p>
            <div className={helpCodeBlockClassName}>Authorization: {authorizationExample}</div>
          </div>

          <div className="space-y-2 text-sm text-slate-600 dark:text-slate-400">
            <p>1. Copy the server URL into your MCP client configuration.</p>
            <p>2. Paste the generated key as a Bearer token in the `Authorization` header.</p>
            <p>3. Keep the key in a secure secret manager or client config, not in shared notes.</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

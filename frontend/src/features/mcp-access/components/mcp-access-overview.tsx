import { Check, Copy, Shield } from 'lucide-react';
import { MCPHelpDialog } from './mcp-help-dialog';
import { Button } from '@/shared/components/button';
import { useCopyFeedback } from '@/shared/hooks/use-copy-feedback';
import { handleSuccess } from '@/shared/utils/handle-success';
import { getMCPServerURL } from '@/features/mcp-access/services/mcp-api-keys';
import { copyToClipboard } from '@/shared/utils/clipboard';

export const MCPAccessOverview = () => {
  const { copiedValue, markCopied } = useCopyFeedback<'url'>();

  const serverUrl = getMCPServerURL();

  const copyServerURL = async () => {
    await copyToClipboard(serverUrl);
    markCopied('url');
    handleSuccess('Server URL copied');
  };

  return (
    <section className="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
      <div className="border-border bg-card rounded-2xl border p-5 shadow-sm">
        <div className="flex items-start gap-3">
          <div className="bg-primary/10 text-primary rounded-xl p-2">
            <Shield className="h-5 w-5" />
          </div>
          <div className="space-y-2">
            <h2 className="text-foreground text-lg font-semibold">What MCP access does</h2>
            <p className="text-muted-foreground text-sm">
              Generate scoped API keys for compatible MCP clients. Each key acts as your signed-in user, so only grant
              the permissions a client actually needs.
            </p>
            <MCPHelpDialog serverUrl={serverUrl} />
          </div>
        </div>
      </div>

      <div className="border-border bg-muted/80 rounded-2xl border p-5 shadow-sm">
        <p className="text-foreground text-sm font-semibold">Server endpoint</p>
        <p className="text-muted-foreground mt-1 text-sm">Configure your external MCP client with this server URL.</p>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
          <div className="border-border bg-card min-w-0 flex-1 rounded-xl border px-3 py-2 font-mono text-sm break-all">
            {serverUrl}
          </div>
          <Button type="button" variant="outline" className="sm:shrink-0" onClick={() => void copyServerURL()}>
            {copiedValue === 'url' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            {copiedValue === 'url' ? 'Copied' : 'Copy URL'}
          </Button>
        </div>
      </div>
    </section>
  );
};

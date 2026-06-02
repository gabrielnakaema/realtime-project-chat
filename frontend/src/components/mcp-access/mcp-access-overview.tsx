import { Check, Copy, Shield } from 'lucide-react';
import { Button } from '../button';
import { MCPHelpDialog } from './mcp-help-dialog';
import { useCopyFeedback } from '@/hooks/use-copy-feedback';
import { handleSuccess } from '@/utils/handle-success';
import { getMCPServerURL } from '@/services/mcp-api-keys';
import { copyToClipboard } from '@/utils/clipboard';

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
      <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-950">
        <div className="flex items-start gap-3">
          <div className="rounded-xl bg-blue-100 p-2 text-blue-700 dark:bg-blue-950 dark:text-blue-300">
            <Shield className="h-5 w-5" />
          </div>
          <div className="space-y-2">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">What MCP access does</h2>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Generate scoped API keys for compatible MCP clients. Each key acts as your signed-in user, so only grant
              the permissions a client actually needs.
            </p>
            <MCPHelpDialog serverUrl={serverUrl} />
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-slate-50/80 p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900/60">
        <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">Server endpoint</p>
        <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
          Configure your external MCP client with this server URL.
        </p>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
          <div className="min-w-0 flex-1 rounded-xl border border-slate-200 bg-white px-3 py-2 font-mono text-sm break-all dark:border-slate-700 dark:bg-slate-950">
            {serverUrl}
          </div>
          <Button type="button" variant="secondary" className="sm:shrink-0" onClick={() => void copyServerURL()}>
            {copiedValue === 'url' ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            {copiedValue === 'url' ? 'Copied' : 'Copy URL'}
          </Button>
        </div>
      </div>
    </section>
  );
};

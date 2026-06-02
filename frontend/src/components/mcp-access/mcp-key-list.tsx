import { useQuery } from '@tanstack/react-query';
import { RefreshCcw, Trash2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Button } from '../button';
import { LoadingSpinner } from '../loading';
import { CreateMCPKeyButton } from './create-mcp-key-dialog';
import { formatMCPAPIKeyLastUsed, getMCPAccessScopeLabel, sortMCPAPIKeys } from './mcp-access-utils';
import { RevokeMCPKeyDialog } from './revoke-mcp-key-dialog';
import type { MCPAPIAvailableScope, MCPAPIKey } from '@/services/mcp-api-keys';
import { formatDateTime } from '@/utils/date';
import { mcpAPIKeyQueryKeys } from '@/services/query-keys';
import { listAvailableMCPAPIScopes, listMCPAPIKeys } from '@/services/mcp-api-keys';
import { cn } from '@/lib/utils';

export const MCPKeyList = () => {
  return (
    <>
      <section className="space-y-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-950">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">Your keys</h2>
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Review active access, identify unused keys, and revoke integrations you no longer trust.
            </p>
          </div>
          <CreateMCPKeyButton />
        </div>

        <MCPKeyListContent />
      </section>
    </>
  );
};

const MCPKeyListContent = () => {
  const keysQuery = useQuery({
    queryKey: mcpAPIKeyQueryKeys.list,
    queryFn: listMCPAPIKeys,
  });
  const scopesQuery = useQuery({
    queryKey: mcpAPIKeyQueryKeys.scopes,
    queryFn: listAvailableMCPAPIScopes,
  });
  const keys = useMemo(() => sortMCPAPIKeys(keysQuery.data ?? []), [keysQuery.data]);
  const availableScopes = scopesQuery.data ?? [];

  const { isFetching, isError, refetch } = keysQuery;

  if (isFetching) {
    return (
      <div className="flex min-h-40 items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 p-6 dark:border-slate-800 dark:bg-slate-900/50">
        <LoadingSpinner size="2rem" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="space-y-3 rounded-2xl border border-red-200 bg-red-50 p-5 dark:border-red-900/80 dark:bg-red-950/30">
        <div className="space-y-1">
          <p className="font-semibold text-red-900 dark:text-red-100">Could not load your MCP keys</p>
          <p className="text-sm text-red-700 dark:text-red-300">
            Retry to fetch the current list and continue managing access.
          </p>
        </div>
        <Button type="button" variant="secondary" onClick={() => refetch()}>
          <RefreshCcw className="h-4 w-4" />
          Retry
        </Button>
      </div>
    );
  }

  if (keys.length === 0) {
    return (
      <div className="rounded-2xl border border-dashed border-slate-300 bg-slate-50/80 p-8 text-center dark:border-slate-700 dark:bg-slate-900/60">
        <p className="font-semibold text-slate-900 dark:text-slate-100">No MCP keys yet</p>
        <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
          Create your first key to connect an external agent or MCP client.
        </p>
        <CreateMCPKeyButton className="mx-auto mt-4" />
      </div>
    );
  }

  return (
    <div className="grid gap-3">
      {keys.map((key) => (
        <MCPKeyCard key={key.id} keyData={key} availableScopes={availableScopes} />
      ))}
    </div>
  );
};

const revokedClassNames = {
  container:
    'border-slate-200 bg-slate-50/70 text-slate-500 dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-400',
  badge: 'bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
};

const activeClassNames = {
  container: 'border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950',
  badge: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300',
};

const MCPKeyCard = ({ keyData, availableScopes }: { keyData: MCPAPIKey; availableScopes: MCPAPIAvailableScope[] }) => {
  const isRevoked = !!keyData.revoked_at;
  const [revokeDialogOpen, setRevokeDialogOpen] = useState(false);

  const classNames = isRevoked ? revokedClassNames : activeClassNames;

  return (
    <>
      <article className={cn('space-y-4 rounded-2xl border p-4 shadow-sm transition-colors', classNames.container)}>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="font-semibold text-slate-900 dark:text-slate-100">{keyData.name}</h3>
              <span className={cn('rounded-full px-2 py-1 text-xs font-medium', classNames.badge)}>
                {isRevoked ? 'Revoked' : 'Active'}
              </span>
            </div>
            <p className="font-mono text-sm">{keyData.key_prefix}</p>
          </div>

          {!isRevoked && (
            <Button
              type="button"
              variant="secondary"
              className="bg-gray-600 text-white hover:bg-gray-700 dark:bg-gray-700 dark:text-white dark:hover:bg-gray-600"
              onClick={() => setRevokeDialogOpen(true)}
            >
              <Trash2 className="h-4 w-4" />
              Revoke
            </Button>
          )}
        </div>

        <div className="grid gap-3 text-sm text-slate-600 md:grid-cols-3 dark:text-slate-400">
          <div>
            <p className="font-medium text-slate-900 dark:text-slate-100">Created</p>
            <p>{formatDateTime(keyData.created_at)}</p>
          </div>
          <div>
            <p className="font-medium text-slate-900 dark:text-slate-100">Last used</p>
            <p>{formatMCPAPIKeyLastUsed(keyData.last_used_at)}</p>
          </div>
          <div>
            <p className="font-medium text-slate-900 dark:text-slate-100">Scopes</p>
            <div className="mt-1 flex flex-wrap gap-2">
              {keyData.scopes.map((scope) => (
                <span
                  key={scope}
                  className="rounded-full border border-slate-200 bg-slate-100 px-2 py-1 text-xs text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300"
                >
                  {getMCPAccessScopeLabel(scope, availableScopes)}
                </span>
              ))}
            </div>
          </div>
        </div>
      </article>

      {!isRevoked && (
        <RevokeMCPKeyDialog isOpen={revokeDialogOpen} keyData={keyData} onOpenChange={setRevokeDialogOpen} />
      )}
    </>
  );
};

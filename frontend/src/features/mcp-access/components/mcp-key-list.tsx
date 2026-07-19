import { useQuery } from '@tanstack/react-query';
import { Pencil, RefreshCcw, Trash2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Button } from '../../../components/button';
import { CreateMCPKeyButton } from './create-mcp-key-dialog';
import { EditMCPKeyDialog } from './edit-mcp-key-dialog';
import { formatMCPAPIKeyLastUsed, getMCPAccessScopeLabel, sortMCPAPIKeys } from './mcp-access-utils';
import { RevokeMCPKeyDialog } from './revoke-mcp-key-dialog';
import type { MCPAPIAvailableScope, MCPAPIKey } from '@/features/mcp-access/services/mcp-api-keys';
import { LoadingSpinner } from '@/shared/components/loading';
import { formatDateTime } from '@/shared/utils/date';
import { mcpAPIKeyQueryKeys } from '@/shared/services/query-keys';
import { listAvailableMCPAPIScopes, listMCPAPIKeys } from '@/features/mcp-access/services/mcp-api-keys';
import { cn } from '@/lib/utils';

export const MCPKeyList = () => {
  return (
    <>
      <section className="border-border bg-card space-y-4 rounded-2xl border p-5 shadow-sm">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-foreground text-lg font-semibold">Your keys</h2>
            <p className="text-muted-foreground text-sm">
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
      <div className="border-border bg-muted flex min-h-40 items-center justify-center rounded-2xl border p-6">
        <LoadingSpinner size="2rem" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="border-destructive/30 bg-destructive/10 space-y-3 rounded-2xl border p-5">
        <div className="space-y-1">
          <p className="text-destructive font-semibold">Could not load your MCP keys</p>
          <p className="text-destructive text-sm">Retry to fetch the current list and continue managing access.</p>
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
      <div className="border-border bg-muted/80 rounded-2xl border border-dashed p-8 text-center">
        <p className="text-foreground font-semibold">No MCP keys yet</p>
        <p className="text-muted-foreground mt-1 text-sm">
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
  container: 'border-border bg-muted/70 text-muted-foreground   ',
  badge: 'bg-muted text-foreground',
};

const activeClassNames = {
  container: 'border-border bg-card  ',
  badge: 'bg-success/10 text-success  ',
};

const MCPKeyCard = ({ keyData, availableScopes }: { keyData: MCPAPIKey; availableScopes: MCPAPIAvailableScope[] }) => {
  const isRevoked = !!keyData.revoked_at;
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [revokeDialogOpen, setRevokeDialogOpen] = useState(false);

  const classNames = isRevoked ? revokedClassNames : activeClassNames;

  return (
    <>
      <article className={cn('space-y-4 rounded-2xl border p-4 shadow-sm transition-colors', classNames.container)}>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-foreground font-semibold">{keyData.name}</h3>
              <span className={cn('rounded-full px-2 py-1 text-xs font-medium', classNames.badge)}>
                {isRevoked ? 'Revoked' : 'Active'}
              </span>
            </div>
            <p className="font-mono text-sm">{keyData.key_prefix}</p>
          </div>

          {!isRevoked && (
            <div className="flex flex-wrap items-center gap-2">
              <Button type="button" variant="secondary" onClick={() => setEditDialogOpen(true)}>
                <Pencil className="h-4 w-4" />
                Edit
              </Button>
              <Button
                type="button"
                variant="secondary"
                className="bg-secondary text-secondary-foreground hover:bg-secondary/80"
                onClick={() => setRevokeDialogOpen(true)}
              >
                <Trash2 className="h-4 w-4" />
                Revoke
              </Button>
            </div>
          )}
        </div>

        <div className="text-muted-foreground grid gap-3 text-sm md:grid-cols-3">
          <div>
            <p className="text-foreground font-medium">Created</p>
            <p>{formatDateTime(keyData.created_at)}</p>
          </div>
          <div>
            <p className="text-foreground font-medium">Last used</p>
            <p>{formatMCPAPIKeyLastUsed(keyData.last_used_at)}</p>
          </div>
          <div>
            <p className="text-foreground font-medium">Scopes</p>
            <div className="mt-1 flex flex-wrap gap-2">
              {keyData.scopes.map((scope) => (
                <span
                  key={scope}
                  className="border-border bg-muted text-foreground rounded-full border px-2 py-1 text-xs"
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
      {!isRevoked && <EditMCPKeyDialog isOpen={editDialogOpen} keyData={keyData} onOpenChange={setEditDialogOpen} />}
    </>
  );
};

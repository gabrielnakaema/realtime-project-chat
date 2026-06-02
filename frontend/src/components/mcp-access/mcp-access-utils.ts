import type { MCPAPIKey, MCPAPIScope } from '@/services/mcp-api-keys';
import { mcpAccessScopeOptions } from '@/services/mcp-api-keys';
import { formatDateTime } from '@/utils/date';

export type CreateModalMode = 'create' | 'reveal';

const scopeOptionsByValue = new Map(mcpAccessScopeOptions.map((scope) => [scope.value, scope]));

export const sortMCPAPIKeys = (keys: MCPAPIKey[]) =>
  [...keys].sort((left, right) => {
    if (!!left.revoked_at !== !!right.revoked_at) {
      return left.revoked_at ? 1 : -1;
    }

    return new Date(right.created_at).getTime() - new Date(left.created_at).getTime();
  });

export const formatMCPAPIKeyLastUsed = (date: string | null) => {
  return formatDateTime(date, 'Never used');
};

export const getMCPAccessScopeLabel = (scope: MCPAPIScope) => {
  return scopeOptionsByValue.get(scope)?.label ?? scope;
};

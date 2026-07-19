import type { MCPAPIAvailableScope, MCPAPIKey, MCPAPIScope } from '@/features/mcp-access/services/mcp-api-keys';
import { formatDateTime } from '@/shared/utils/date';

export type CreateModalMode = 'create' | 'reveal';

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

export const getMCPAccessScopeLabel = (scope: MCPAPIScope, availableScopes: MCPAPIAvailableScope[] = []) => {
  return availableScopes.find((option) => option.scope === scope)?.label ?? scope;
};

import { useQuery } from '@tanstack/react-query';
import { listAvailableMCPAPIScopes } from '@/features/mcp-access/services/mcp-api-keys';
import { mcpAPIKeyQueryKeys } from '@/shared/services/query-keys';

export const useMCPAPIAvailableScopes = () => {
  const query = useQuery({
    queryKey: mcpAPIKeyQueryKeys.scopes,
    queryFn: listAvailableMCPAPIScopes,
  });

  const availableScopes = query.data ?? [];

  return {
    availableScopes,
    isLoadingScopes: query.isLoading && availableScopes.length === 0,
    isScopeLoadFailed: query.isError && availableScopes.length === 0,
    refetchScopes: query.refetch,
  };
};

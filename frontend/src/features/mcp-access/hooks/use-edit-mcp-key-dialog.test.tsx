// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useEditMcpKeyDialog } from './use-edit-mcp-key-dialog';
import type { ReactNode } from 'react';
import type { IMCPAPIKeyForm } from '@/features/mcp-access/schemas/mcp-api-key.schema';
import { updateMCPAPIKey } from '@/features/mcp-access/services/mcp-api-keys';
import { mcpAPIKeyQueryKeys } from '@/shared/services/query-keys';
import { handleSuccess } from '@/shared/utils/handle-success';

vi.mock('@/features/mcp-access/services/mcp-api-keys', () => ({
  updateMCPAPIKey: vi.fn(),
}));

vi.mock('@/shared/utils/handle-success', () => ({
  handleSuccess: vi.fn(),
}));

const mockUpdateMCPAPIKey = vi.mocked(updateMCPAPIKey);
const mockHandleSuccess = vi.mocked(handleSuccess);

const existingKey = {
  id: 'key-1',
  name: 'Editor key',
  key_prefix: 'mcp_123',
  scopes: ['tasks:move'],
  created_at: '2026-06-02T10:00:00.000Z',
  last_used_at: null,
  revoked_at: null,
};

const updatedKey = {
  ...existingKey,
  name: 'Updated editor key',
  scopes: ['tasks:read'],
};

const createWrapper = (queryClient: QueryClient) => {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useEditMcpKeyDialog', () => {
  beforeEach(() => {
    mockUpdateMCPAPIKey.mockResolvedValue(updatedKey);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('preloads default values from the provided key', () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    const { result } = renderHook(
      () =>
        useEditMcpKeyDialog({
          keyData: existingKey,
          onOpenChange: vi.fn(),
        }),
      {
        wrapper: createWrapper(queryClient),
      },
    );

    expect(result.current.defaultValues).toEqual({
      name: 'Editor key',
      scopes: ['tasks:move'],
    });
  });

  it('updates the cache and closes the dialog after a successful save', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    const onOpenChange = vi.fn();
    queryClient.setQueryData(mcpAPIKeyQueryKeys.list, [existingKey]);

    const { result } = renderHook(
      () =>
        useEditMcpKeyDialog({
          keyData: existingKey,
          onOpenChange,
        }),
      {
        wrapper: createWrapper(queryClient),
      },
    );

    act(() => {
      result.current.submitUpdate({
        name: 'Updated editor key',
        scopes: ['tasks:read'],
      } satisfies IMCPAPIKeyForm);
    });

    await waitFor(() => {
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });

    expect(mockUpdateMCPAPIKey).toHaveBeenCalledWith('key-1', {
      name: 'Updated editor key',
      scopes: ['tasks:read'],
    });
    expect(mockHandleSuccess).toHaveBeenCalledWith('Key updated');
    expect(queryClient.getQueryData(mcpAPIKeyQueryKeys.list)).toEqual([updatedKey]);
  });
});

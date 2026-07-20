// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useCreateMcpKeyDialog } from './use-create-mcp-key-dialog';
import type { ReactNode } from 'react';
import { mcpAPIKeyQueryKeys } from '@/shared/services/query-keys';
import { copyToClipboard } from '@/shared/utils/clipboard';
import { handleSuccess } from '@/shared/utils/handle-success';
import { createMCPAPIKey, getMCPServerURL } from '@/features/mcp-access/services/mcp-api-keys';

vi.mock('@/features/mcp-access/services/mcp-api-keys', () => ({
  createMCPAPIKey: vi.fn(),
  getMCPServerURL: vi.fn(),
}));

vi.mock('@/shared/utils/clipboard', () => ({
  copyToClipboard: vi.fn(),
}));

vi.mock('@/shared/utils/handle-success', () => ({
  handleSuccess: vi.fn(),
}));

const mockCreateMCPAPIKey = vi.mocked(createMCPAPIKey);
const mockGetMCPServerURL = vi.mocked(getMCPServerURL);
const mockCopyToClipboard = vi.mocked(copyToClipboard);
const mockHandleSuccess = vi.mocked(handleSuccess);

const existingKey = {
  id: 'key-existing',
  name: 'Existing key',
  key_prefix: 'mcp_existing',
  scopes: ['tasks:read'],
  created_at: '2026-06-01T10:00:00.000Z',
  last_used_at: null,
  revoked_at: null,
};

const createdResponse = {
  key: {
    id: 'key-new',
    name: 'Fresh key',
    key_prefix: 'mcp_new',
    scopes: ['tasks:move'],
    created_at: '2026-06-05T10:00:00.000Z',
    last_used_at: null,
    revoked_at: null,
  },
  raw_secret: 'secret-value',
};

const createWrapper = (queryClient: QueryClient) => {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useCreateMcpKeyDialog', () => {
  beforeEach(() => {
    mockGetMCPServerURL.mockReturnValue('https://example.com/mcp');
    mockCreateMCPAPIKey.mockResolvedValue(createdResponse);
    mockCopyToClipboard.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('transitions into reveal mode after a successful create and updates the cache', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    queryClient.setQueryData(mcpAPIKeyQueryKeys.list, [existingKey]);

    const { result } = renderHook(() => useCreateMcpKeyDialog(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.handleDialogOpenChange(true);
      result.current.createKey({ name: 'Fresh key', scopes: ['tasks:move'] });
    });

    await waitFor(() => {
      expect(result.current.isRevealMode).toBe(true);
    });

    expect(result.current.open).toBe(true);
    expect(result.current.rawSecret).toBe('secret-value');
    expect(result.current.serverUrl).toBe('https://example.com/mcp');
    expect(queryClient.getQueryData(mcpAPIKeyQueryKeys.list)).toEqual([createdResponse.key, existingKey]);
  });

  it('shows a confirmation prompt when Close is clicked before copying the secret', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    const { result } = renderHook(() => useCreateMcpKeyDialog(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.handleDialogOpenChange(true);
      result.current.createKey({ name: 'Fresh key', scopes: ['tasks:move'] });
    });

    await waitFor(() => {
      expect(result.current.isRevealMode).toBe(true);
    });

    // Simulate the Close button click (routes through handleDialogOpenChange)
    act(() => {
      result.current.handleDialogOpenChange(false);
    });

    expect(result.current.open).toBe(true);
    expect(result.current.dismissRevealPromptOpen).toBe(true);
  });

  it('allows closing without a prompt after the secret has been copied', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    const { result } = renderHook(() => useCreateMcpKeyDialog(), {
      wrapper: createWrapper(queryClient),
    });

    act(() => {
      result.current.handleDialogOpenChange(true);
      result.current.createKey({ name: 'Fresh key', scopes: ['tasks:move'] });
    });

    await waitFor(() => {
      expect(result.current.isRevealMode).toBe(true);
    });

    await act(async () => {
      await result.current.copyValue('secret-value', 'secret');
    });

    expect(mockCopyToClipboard).toHaveBeenCalledWith('secret-value');
    expect(mockHandleSuccess).toHaveBeenCalledWith('Secret copied');

    act(() => {
      result.current.handleDialogOpenChange(false);
    });

    expect(result.current.dismissRevealPromptOpen).toBe(false);
    expect(result.current.open).toBe(false);
    expect(result.current.isRevealMode).toBe(false);
    expect(result.current.rawSecret).toBe('');
  });
});

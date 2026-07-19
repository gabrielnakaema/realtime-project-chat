import { describe, expect, it } from 'vitest';
import { formatMCPAPIKeyLastUsed, getMCPAccessScopeLabel, sortMCPAPIKeys } from './mcp-access-utils';
import type { MCPAPIKey } from '@/features/mcp-access/services/mcp-api-keys';
import { formatDateTime } from '@/shared/utils/date';

const buildKey = (overrides: Partial<MCPAPIKey>): MCPAPIKey => ({
  id: overrides.id ?? 'key-1',
  name: overrides.name ?? 'Claude Desktop',
  key_prefix: overrides.key_prefix ?? 'mcp_123',
  scopes: overrides.scopes ?? ['tasks:read'],
  created_at: overrides.created_at ?? '2026-06-01T12:00:00.000Z',
  last_used_at: overrides.last_used_at ?? null,
  revoked_at: overrides.revoked_at ?? null,
});

describe('mcp access utils', () => {
  it('sorts active keys before revoked keys and newer keys first', () => {
    const sorted = sortMCPAPIKeys([
      buildKey({ id: 'revoked', revoked_at: '2026-06-02T12:00:00.000Z' }),
      buildKey({ id: 'older-active', created_at: '2026-05-01T12:00:00.000Z' }),
      buildKey({ id: 'newer-active', created_at: '2026-06-02T12:00:00.000Z' }),
    ]);

    expect(sorted.map((key) => key.id)).toEqual(['newer-active', 'older-active', 'revoked']);
  });

  it('formats last-used dates with the shared fallback', () => {
    expect(formatMCPAPIKeyLastUsed(null)).toBe('Never used');
    expect(formatMCPAPIKeyLastUsed('2026-06-02T15:45:00.000Z')).toBe(formatDateTime('2026-06-02T15:45:00.000Z'));
  });

  it('falls back to the raw scope when the scope catalog has not loaded it yet', () => {
    expect(
      getMCPAccessScopeLabel('tasks:review', [{ scope: 'tasks:read', label: 'Read tasks', title: 'Read tasks' }]),
    ).toBe('tasks:review');
  });
});

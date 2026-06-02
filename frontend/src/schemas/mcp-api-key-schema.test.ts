import { describe, expect, it } from 'vitest';
import { createMCPAPIKeySchema } from './mcp-api-key-schema';

describe('mcp api key schema', () => {
  it('requires a non-empty name after trimming', () => {
    const result = createMCPAPIKeySchema.safeParse({
      name: '   ',
      scopes: ['tasks:read'],
    });

    expect(result.success).toBe(false);
    expect(result.error?.issues[0]?.message).toBe('Name is required');
  });

  it('requires at least one scope', () => {
    const result = createMCPAPIKeySchema.safeParse({
      name: 'Claude Desktop',
      scopes: [],
    });

    expect(result.success).toBe(false);
    expect(result.error?.issues[0]?.message).toBe('Select at least one scope');
  });

  it('returns trimmed values for valid payloads', () => {
    const result = createMCPAPIKeySchema.safeParse({
      name: '  Claude Desktop  ',
      scopes: ['tasks:read', 'tasks:comment'],
    });

    expect(result.success).toBe(true);
    expect(result.data).toEqual({
      name: 'Claude Desktop',
      scopes: ['tasks:read', 'tasks:comment'],
    });
  });
});

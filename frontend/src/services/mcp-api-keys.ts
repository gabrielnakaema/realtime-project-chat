import { api, baseApiUrl } from './api';

export type MCPAPIScope = string;

export interface MCPAPIKey {
  id: string;
  name: string;
  key_prefix: string;
  scopes: MCPAPIScope[];
  created_at: string;
  last_used_at: string | null;
  revoked_at: string | null;
}

export interface CreateMCPAPIKeyPayload {
  name: string;
  scopes: MCPAPIScope[];
}

export interface CreateMCPAPIKeyResponse {
  key: MCPAPIKey;
  raw_secret: string;
}

export interface MCPAPIAvailableScope {
  scope: MCPAPIScope;
  label: string;
  title: string;
}

export const listMCPAPIKeys = async () => {
  const response = await api.get('users/me/mcp-api-keys');
  return response.json<MCPAPIKey[]>();
};

export const listAvailableMCPAPIScopes = async () => {
  const response = await api.get('users/me/mcp-api-keys/scopes');
  return response.json<MCPAPIAvailableScope[]>();
};

export const createMCPAPIKey = async (payload: CreateMCPAPIKeyPayload) => {
  const response = await api.post('users/me/mcp-api-keys', {
    json: payload,
  });

  return response.json<CreateMCPAPIKeyResponse>();
};

export const revokeMCPAPIKey = async (id: string) => {
  await api.delete(`users/me/mcp-api-keys/${id}`);
};

export const getMCPServerURL = () => {
  if (!baseApiUrl) {
    return '/mcp';
  }

  try {
    const url = new URL(baseApiUrl);
    return `${url.origin}/mcp`;
  } catch {
    return `${baseApiUrl.replace(/\/+$/, '')}/mcp`;
  }
};

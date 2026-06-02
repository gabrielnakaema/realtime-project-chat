import { api, baseApiUrl } from './api';

export const mcpApiScopeValues = [
  'projects:read',
  'tasks:read',
  'tasks:move',
  'tasks:comment',
  'tasks:mark_done',
  'tasks:assign:self',
] as const;

export type MCPAPIScope = (typeof mcpApiScopeValues)[number];

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

export interface MCPAccessScopeOption {
  value: MCPAPIScope;
  label: string;
  description: string;
}

export const mcpAccessScopeOptions: MCPAccessScopeOption[] = [
  {
    value: 'projects:read',
    label: 'Read projects',
    description: 'View project details and browse the workspace structure.',
  },
  {
    value: 'tasks:read',
    label: 'Read tasks',
    description: 'Inspect task fields, status, assignments, and other task metadata.',
  },
  {
    value: 'tasks:move',
    label: 'Move tasks',
    description: 'Change where a task sits in the workflow.',
  },
  {
    value: 'tasks:comment',
    label: 'Comment on tasks',
    description: 'Post task comments on your behalf from an MCP client.',
  },
  {
    value: 'tasks:mark_done',
    label: 'Mark tasks done',
    description: 'Complete tasks by moving them into a done state when supported.',
  },
  {
    value: 'tasks:assign:self',
    label: 'Assign tasks to me',
    description: 'Let an MCP client assign tasks to your own account only.',
  },
];

export const listMCPAPIKeys = async () => {
  const response = await api.get('users/me/mcp-api-keys');
  return response.json<MCPAPIKey[]>();
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

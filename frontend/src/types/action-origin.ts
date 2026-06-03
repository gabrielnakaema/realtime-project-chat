export type ActionOrigin = 'user' | 'mcp_agent';

export const normalizeActionOrigin = (origin?: ActionOrigin | null): ActionOrigin => {
  if (origin === 'mcp_agent') {
    return 'mcp_agent';
  }

  return 'user';
};

export const isMCPAgentAction = (origin?: ActionOrigin | null): boolean => {
  return normalizeActionOrigin(origin) === 'mcp_agent';
};

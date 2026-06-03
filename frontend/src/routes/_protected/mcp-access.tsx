import { createFileRoute } from '@tanstack/react-router';
import { MCPAccessPage } from '@/components/mcp-access/mcp-access-page';

export const Route = createFileRoute('/_protected/mcp-access')({
  component: MCPAccessPage,
});

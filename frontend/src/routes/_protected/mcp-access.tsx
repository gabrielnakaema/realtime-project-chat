import { createFileRoute } from '@tanstack/react-router';
import { MCPAccessPage } from '@/features/mcp-access/components/mcp-access-page';

export const Route = createFileRoute('/_protected/mcp-access')({
  component: MCPAccessPage,
});

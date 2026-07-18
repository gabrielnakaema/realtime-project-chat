import { Link } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';
import { HeaderUser } from '../header-user';
import { NotificationBell } from '../notification-bell';
import { MCPAccessOverview } from './mcp-access-overview';
import { MCPKeyList } from './mcp-key-list';

export const MCPAccessPage = () => {
  return (
    <div className="bg-muted min-h-screen">
      <header className="border-border bg-card border-b">
        <div className="px-6 py-4">
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="flex items-center gap-4">
              <Link
                to="/projects"
                className="text-foreground hover:bg-muted inline-flex items-center rounded-md px-3 py-2 font-medium whitespace-nowrap transition-colors"
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Go back
              </Link>
              <div>
                <h1 className="text-foreground text-2xl font-bold">MCP Access</h1>
                <p className="text-muted-foreground">
                  Manage agent access and secure API keys for external MCP clients.
                </p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <NotificationBell />
              <HeaderUser />
            </div>
          </div>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-6 py-10">
        <MCPAccessOverview />

        <MCPKeyList />
      </main>
    </div>
  );
};

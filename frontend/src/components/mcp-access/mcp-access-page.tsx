import { Link } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';
import { HeaderUser } from '../header-user';
import { NotificationBell } from '../notification-bell';
import { MCPAccessOverview } from './mcp-access-overview';
import { MCPKeyList } from './mcp-key-list';

export const MCPAccessPage = () => {
  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <div className="px-6 py-4">
          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div className="flex items-center gap-4">
              <Link
                to="/projects"
                className="inline-flex items-center rounded-md px-3 py-2 font-medium whitespace-nowrap text-slate-700 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Go back
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">MCP Access</h1>
                <p className="text-slate-600 dark:text-slate-400">
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

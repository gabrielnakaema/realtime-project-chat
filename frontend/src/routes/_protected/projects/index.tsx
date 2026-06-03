import { createFileRoute } from '@tanstack/react-router';
import { MessageSquare } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { CreateProject } from '@/components/create-project';
import { HeaderUser } from '@/components/header-user';
import { NotificationBell } from '@/components/notification-bell';
import { UnreadCountBadge } from '@/components/unread-count-badge';
import { ProjectList } from '@/components/project-list';
import { SearchBar } from '@/components/search-bar';
import { UserDueTasks } from '@/components/user-due-tasks';
import { UserProjectActivities } from '@/components/user-project-activities';
import { useMessagesSheet } from '@/contexts/messages-sheet-context';
import { useAuth } from '@/hooks/use-auth';
import { listGeneralChats } from '@/services/general-chat';
import { generalChatQueryKeys } from '@/services/query-keys';

export const Route = createFileRoute('/_protected/projects/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = useAuth();
  const { open: openMessages } = useMessagesSheet();
  const { data: generalChats = [] } = useQuery({
    queryKey: generalChatQueryKeys.list,
    queryFn: listGeneralChats,
  });
  const unreadCount = generalChats.reduce((sum, chat) => sum + chat.unread_count, 0);
  const hasMoreUnread = generalChats.some((chat) => chat.has_more_unread) || unreadCount > 99;

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">TaskFlow</h1>
              <p className="text-slate-600 dark:text-slate-400">Welcome back, {user?.name}</p>
            </div>

            <SearchBar action="/search" searchName="query" formClassName="w-full max-w-md" />

            <div className="flex items-center gap-4">
              <button
                onClick={openMessages}
                className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
              >
                <MessageSquare className="h-4 w-4" />
                Messages
                <UnreadCountBadge count={unreadCount} hasMoreUnread={hasMoreUnread} />
              </button>

              <NotificationBell />

              <CreateProject />

              <HeaderUser />
            </div>
          </div>
        </div>
      </header>

      <div className="space-y-8 px-6 py-12">
        <ProjectList />

        <UserDueTasks />

        <UserProjectActivities />
      </div>
    </div>
  );
}

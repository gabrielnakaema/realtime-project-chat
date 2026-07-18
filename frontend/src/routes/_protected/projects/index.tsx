import { createFileRoute } from '@tanstack/react-router';
import { MessageSquare } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { CreateProject } from '@/components/project-form/create-project';
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
    <div className="bg-muted min-h-screen">
      <header className="border-border bg-card border-b">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-foreground text-2xl font-bold">TaskFlow</h1>
              <p className="text-muted-foreground">Welcome back, {user?.name}</p>
            </div>

            <SearchBar action="/search" searchName="query" formClassName="w-full max-w-md" />

            <div className="flex items-center gap-4">
              <button
                onClick={openMessages}
                className="text-foreground hover:bg-muted flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors"
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

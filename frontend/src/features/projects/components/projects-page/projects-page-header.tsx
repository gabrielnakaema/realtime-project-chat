import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { MessageSquare, Search } from 'lucide-react';
import { HeaderUser } from '@/components/header-user';
import { NotificationBell } from '@/components/notification-bell';
import { SearchBar } from '@/components/search-bar';
import { UnreadCountBadge } from '@/components/unread-count-badge';
import { useMessagesSheet } from '@/contexts/messages-sheet-context';
import { listGeneralChats } from '@/services/general-chat';
import { generalChatQueryKeys } from '@/services/query-keys';
import { HeaderActions, HeaderCenter, HeaderLogo, HeaderShell, HeaderStart } from '@/shared/components/header-shell';

export const ProjectsPageHeader = () => {
  const { open: openMessages } = useMessagesSheet();
  const { data: generalChats = [] } = useQuery({
    queryKey: generalChatQueryKeys.list,
    queryFn: listGeneralChats,
  });
  const unreadCount = generalChats.reduce((sum, chat) => sum + chat.unread_count, 0);
  const hasMoreUnread = generalChats.some((chat) => chat.has_more_unread) || unreadCount > 99;

  return (
    <HeaderShell width="contained">
      <HeaderStart>
        <HeaderLogo />
        <span className="text-foreground text-sm font-semibold tracking-[-0.01em]">TaskFlow</span>
      </HeaderStart>

      <HeaderCenter className="hidden justify-center sm:flex">
        <SearchBar action="/search" searchName="query" formClassName="w-full max-w-sm" />
      </HeaderCenter>

      <HeaderActions>
        <Link
          to="/search"
          className="text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-ring flex size-9 items-center justify-center rounded-lg transition-colors focus-visible:ring-2 focus-visible:outline-none sm:hidden"
          aria-label="Search projects and tasks"
        >
          <Search className="size-4" />
        </Link>

        <button
          type="button"
          onClick={openMessages}
          className="text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-ring relative flex size-9 items-center justify-center rounded-lg transition-colors focus-visible:ring-2 focus-visible:outline-none"
          aria-label="Messages"
        >
          <MessageSquare className="size-[15px]" />
          <UnreadCountBadge
            count={unreadCount}
            hasMoreUnread={hasMoreUnread}
            className="absolute -top-1 -right-1 min-w-4 px-1 py-0 text-[9px] leading-4"
          />
        </button>

        <NotificationBell />
        <HeaderUser />
      </HeaderActions>
    </HeaderShell>
  );
};

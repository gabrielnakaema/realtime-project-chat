import { useQuery } from '@tanstack/react-query';
import { PencilLine, X } from 'lucide-react';
import { ComposeModal } from './compose-modal';
import { getAvatarColorClass, getChatAvatarSeed, getChatSubtitle, getChatTitle } from './utils';
import { UnreadCountBadge } from '@/features/chat/components/unread-count-badge';
import { useMessagesSheet } from '@/features/chat/components/messages-sheet-context';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { cn } from '@/lib/utils';
import { listGeneralChats } from '@/features/chat/services/general-chat';
import { generalChatQueryKeys } from '@/shared/services/query-keys';

export const MessagesListView = () => {
  const { user } = useAuth();
  const { openChat, openCompose, close, view } = useMessagesSheet();

  const { data: chats = [] } = useQuery({
    queryKey: generalChatQueryKeys.list,
    queryFn: listGeneralChats,
  });

  return (
    <>
      <div className="border-border flex shrink-0 items-center justify-between border-b px-4 py-3">
        <h2 className="text-foreground text-base font-semibold">Messages</h2>
        <div className="flex items-center gap-1">
          <button
            onClick={openCompose}
            className="text-muted-foreground hover:bg-muted hover:text-foreground flex h-8 w-8 items-center justify-center rounded-lg transition-colors"
            title="New message"
          >
            <PencilLine className="h-4 w-4" />
          </button>
          <button
            onClick={close}
            className="text-muted-foreground hover:bg-muted hover:text-muted-foreground flex h-8 w-8 items-center justify-center rounded-lg transition-colors"
            title="Close"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {chats.length === 0 ? (
          <div className="flex flex-col items-center justify-center px-6 py-20 text-center">
            <div className="bg-muted mb-3 flex h-12 w-12 items-center justify-center rounded-2xl">
              <PencilLine className="text-muted-foreground h-5 w-5" />
            </div>
            <p className="text-foreground text-sm font-medium">No messages yet</p>
            <p className="text-muted-foreground mt-1 text-xs">Start a conversation with a teammate</p>
            <button
              onClick={openCompose}
              className="bg-primary text-primary-foreground hover:bg-primary/90 mt-4 rounded-lg px-4 py-1.5 text-xs font-medium transition-colors"
            >
              New message
            </button>
          </div>
        ) : (
          <ul className="px-2 py-2">
            {chats.map((chat) => (
              <li key={chat.id}>
                <button
                  onClick={() => openChat(chat.id)}
                  className="hover:bg-muted flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left transition-colors"
                >
                  <div
                    className={cn(
                      'flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white',
                      getAvatarColorClass(getChatAvatarSeed(chat, user?.id)),
                    )}
                  >
                    {getChatAvatarSeed(chat, user?.id).charAt(0).toUpperCase()}
                  </div>
                  <div className="min-w-0">
                    <p className="text-foreground truncate text-sm font-medium">{getChatTitle(chat, user?.id)}</p>
                    <div className="flex items-center gap-2">
                      <p className="text-muted-foreground text-xs">{getChatSubtitle(chat, user?.id)}</p>
                      <UnreadCountBadge count={chat.unread_count} hasMoreUnread={chat.has_more_unread} />
                    </div>
                  </div>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {view === 'compose' && <ComposeModal />}
    </>
  );
};

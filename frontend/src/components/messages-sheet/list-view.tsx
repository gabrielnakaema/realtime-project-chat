import { useQuery } from '@tanstack/react-query';
import { PencilLine, X } from 'lucide-react';
import { useState } from 'react';
import { MessagesComposeModal } from './compose-modal';
import { getAvatarColorClass, getChatAvatarSeed, getChatSubtitle, getChatTitle } from './utils';
import { useMessagesSheet } from '@/contexts/messages-sheet-context';
import { useAuth } from '@/hooks/use-auth';
import { cn } from '@/lib/utils';
import { listGeneralChats } from '@/services/general-chat';
import { generalChatQueryKeys } from '@/services/query-keys';

export function MessagesListView() {
  const { user } = useAuth();
  const { openChat, close } = useMessagesSheet();
  const [showCompose, setShowCompose] = useState(false);

  const { data: chats = [] } = useQuery({
    queryKey: generalChatQueryKeys.list,
    queryFn: listGeneralChats,
  });

  return (
    <>
      <div className="flex shrink-0 items-center justify-between border-b border-slate-100 px-4 py-3 dark:border-slate-700">
        <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">Messages</h2>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setShowCompose(true)}
            className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800 dark:hover:text-slate-300"
            title="New message"
          >
            <PencilLine className="h-4 w-4" />
          </button>
          <button
            onClick={close}
            className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-800 dark:hover:text-slate-300"
            title="Close"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {chats.length === 0 ? (
          <div className="flex flex-col items-center justify-center px-6 py-20 text-center">
            <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-slate-100 dark:bg-slate-800">
              <PencilLine className="h-5 w-5 text-slate-400" />
            </div>
            <p className="text-sm font-medium text-slate-700 dark:text-slate-300">No messages yet</p>
            <p className="mt-1 text-xs text-slate-400">Start a conversation with a teammate</p>
            <button
              onClick={() => setShowCompose(true)}
              className="mt-4 rounded-lg bg-blue-600 px-4 py-1.5 text-xs font-medium text-white transition-colors hover:bg-blue-700"
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
                  className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left transition-colors hover:bg-slate-50 dark:hover:bg-slate-800"
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
                    <p className="truncate text-sm font-medium text-slate-900 dark:text-slate-100">
                      {getChatTitle(chat, user?.id)}
                    </p>
                    <p className="text-xs text-slate-400">{getChatSubtitle(chat, user?.id)}</p>
                  </div>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {showCompose && <MessagesComposeModal onClose={() => setShowCompose(false)} />}
    </>
  );
}

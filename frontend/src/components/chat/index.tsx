import { ChatView } from './chat-view';
import { MessagesListView } from './list-view';
import { useMessagesSheet } from '@/contexts/messages-sheet-context';

export const MessagesSheet = () => {
  const { isOpen, selectedChatId, view, close } = useMessagesSheet();

  if (!isOpen) return null;

  return (
    <>
      <div className="fixed inset-0 z-40 bg-black/25 backdrop-blur-[2px]" onClick={close} />
      <aside className="fixed top-0 right-0 z-50 flex h-full w-full max-w-[420px] flex-col border-l border-slate-200 bg-white shadow-2xl dark:border-slate-700 dark:bg-slate-900">
        {view === 'chat' && selectedChatId ? <ChatView chatId={selectedChatId} /> : <MessagesListView />}
      </aside>
    </>
  );
};

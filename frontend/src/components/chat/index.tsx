import { ChatView } from './chat-view';
import { MessagesListView } from './list-view';
import { useMessagesSheet } from '@/contexts/messages-sheet-context';

export const MessagesSheet = () => {
  const { isOpen, selectedChatId, view, close } = useMessagesSheet();

  if (!isOpen) return null;

  return (
    <>
      <div className="bg-overlay/25 fixed inset-0 z-40 backdrop-blur-[2px]" onClick={close} />
      <aside className="border-border bg-card fixed top-0 right-0 z-50 flex h-full w-full max-w-[420px] flex-col border-l shadow-2xl">
        {view === 'chat' && selectedChatId ? <ChatView chatId={selectedChatId} /> : <MessagesListView />}
      </aside>
    </>
  );
};

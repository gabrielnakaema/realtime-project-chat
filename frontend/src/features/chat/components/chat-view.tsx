import { ChatComposer } from './chat-composer';
import { ChatHeader } from './chat-header';
import { ChatMessageList } from './chat-message-list';
import { useMessagesSheet } from '@/features/chat/components/messages-sheet-context';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useGeneralChat } from '@/features/chat/hooks/use-general-chat';

interface ChatViewProps {
  chatId: string;
}

export const ChatView = ({ chatId }: ChatViewProps) => {
  const { user } = useAuth();
  const { backToList, close } = useMessagesSheet();
  const { chat, messages, observedRef, chatContainerRef } = useGeneralChat(chatId);

  return (
    <>
      <ChatHeader chat={chat} currentUserId={user?.id} onBack={backToList} onClose={close} />
      <ChatMessageList
        messages={messages}
        currentUserId={user?.id}
        chatContainerRef={chatContainerRef}
        observedRef={observedRef}
      />
      <ChatComposer chatId={chatId} />
    </>
  );
};

import { MessagesChatComposer } from './chat-composer';
import { MessagesChatHeader } from './chat-header';
import { MessagesChatMessageList } from './chat-message-list';
import { useMessagesSheet } from '@/contexts/messages-sheet-context';
import { useAuth } from '@/hooks/use-auth';
import { useGeneralChat } from '@/hooks/use-general-chat';

interface MessagesChatViewProps {
  chatId: string;
}

export const MessagesChatView = ({ chatId }: MessagesChatViewProps) => {
  const { user } = useAuth();
  const { backToList, close } = useMessagesSheet();
  const { chat, messages, observedRef, chatContainerRef } = useGeneralChat(chatId);

  return (
    <>
      <MessagesChatHeader chat={chat} currentUserId={user?.id} onBack={backToList} onClose={close} />
      <MessagesChatMessageList
        messages={messages}
        currentUserId={user?.id}
        chatContainerRef={chatContainerRef}
        observedRef={observedRef}
      />
      <MessagesChatComposer chatId={chatId} />
    </>
  );
};

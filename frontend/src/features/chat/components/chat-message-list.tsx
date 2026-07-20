import { buildMessagesTimeline } from './chat-timeline';
import { ChatMessageItem } from './chat-message-item';
import type { RefCallback, RefObject } from 'react';
import type { ChatMessage } from '@/features/chat/types/chat';

interface ChatMessageListProps {
  messages: ChatMessage[];
  currentUserId?: string;
  chatContainerRef: RefObject<HTMLDivElement | null>;
  observedRef: RefCallback<HTMLDivElement>;
}

export const ChatMessageList = ({ messages, currentUserId, chatContainerRef, observedRef }: ChatMessageListProps) => {
  const timeline = buildMessagesTimeline(messages, currentUserId);

  return (
    <div ref={chatContainerRef} className="flex-1 overflow-y-auto p-4">
      <div ref={observedRef} className="h-1" />
      <div className="space-y-3">
        {timeline.map((item) => {
          if (item.type === 'date') {
            return (
              <div key={item.id} className="my-4 flex justify-center">
                <span className="bg-muted text-muted-foreground rounded-full px-3 py-0.5 text-xs">{item.label}</span>
              </div>
            );
          }

          return (
            <ChatMessageItem
              key={item.id}
              message={item.message}
              isCurrentUser={item.isCurrentUser}
              currentUserId={currentUserId}
            />
          );
        })}
      </div>
    </div>
  );
};

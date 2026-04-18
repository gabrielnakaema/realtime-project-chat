import { buildChatTimeline } from './chat-timeline';
import { ChatMessageItem } from './chat-message-item';
import type { RefObject } from 'react';
import type { ChatMessage } from '@/types/chat';

interface ChatMessageListProps {
  messages: ChatMessage[];
  currentUserId?: string;
  chatContainerRef: RefObject<HTMLDivElement | null>;
  observedRef: RefObject<HTMLDivElement | null>;
}

export const ChatMessageList = ({ messages, currentUserId, chatContainerRef, observedRef }: ChatMessageListProps) => {
  const timeline = buildChatTimeline(messages, currentUserId);

  return (
    <div ref={chatContainerRef} className="flex-1 overflow-y-auto bg-[var(--background)] p-6">
      <div ref={observedRef} className="h-1 bg-transparent" />
      <div className="mx-auto max-w-4xl space-y-4">
        {timeline.map((item) => {
          if (item.type === 'date') {
            return (
              <div key={item.id} className="my-6 flex justify-center">
                <span className="rounded-full bg-slate-200 px-3 py-1 text-xs text-slate-600 dark:bg-slate-700 dark:text-slate-400">
                  {item.label}
                </span>
              </div>
            );
          }

          return <ChatMessageItem key={item.id} message={item.message} isCurrentUser={item.isCurrentUser} />;
        })}
      </div>
    </div>
  );
};

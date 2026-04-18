import { buildMessagesTimeline } from './chat-timeline';
import { MessagesChatMessageItem } from './chat-message-item';
import type { RefObject } from 'react';
import type { ChatMessage } from '@/types/chat';

interface MessagesChatMessageListProps {
  messages: ChatMessage[];
  currentUserId?: string;
  chatContainerRef: RefObject<HTMLDivElement | null>;
  observedRef: RefObject<HTMLDivElement | null>;
}

export function MessagesChatMessageList({
  messages,
  currentUserId,
  chatContainerRef,
  observedRef,
}: MessagesChatMessageListProps) {
  const timeline = buildMessagesTimeline(messages, currentUserId);

  return (
    <div ref={chatContainerRef} className="flex-1 overflow-y-auto p-4">
      <div ref={observedRef} className="h-1" />
      <div className="space-y-3">
        {timeline.map((item) => {
          if (item.type === 'date') {
            return (
              <div key={item.id} className="my-4 flex justify-center">
                <span className="rounded-full bg-slate-100 px-3 py-0.5 text-xs text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                  {item.label}
                </span>
              </div>
            );
          }

          return <MessagesChatMessageItem key={item.id} message={item.message} isCurrentUser={item.isCurrentUser} />;
        })}
      </div>
    </div>
  );
}

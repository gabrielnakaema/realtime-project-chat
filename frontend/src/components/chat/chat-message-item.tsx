import { useState } from 'react';
import { MoreVertical } from 'lucide-react';
import { formatTime, getAvatarColorClass } from './utils';
import type { ChatMessage } from '@/types/chat';
import { ChatMessageReadsSheet } from '@/components/chat-message-reads-sheet';
import { cn } from '@/lib/utils';

interface ChatMessageItemProps {
  message: ChatMessage;
  isCurrentUser: boolean;
  currentUserId?: string;
}

export const ChatMessageItem = ({ message, isCurrentUser, currentUserId }: ChatMessageItemProps) => {
  const [isReadsSheetOpen, setIsReadsSheetOpen] = useState(false);
  const authorName = message.member?.user?.name ?? '';
  const messageTime = formatTime(message.created_at);
  const avatarInitial = authorName.charAt(0).toUpperCase();
  const bubbleClassName = cn(
    'rounded-2xl px-3 py-2 text-left text-sm leading-relaxed',
    isCurrentUser
      ? 'rounded-tr-sm bg-blue-600 text-white'
      : 'rounded-tl-sm border border-slate-200 bg-white text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100',
  );
  const detailsButtonClassName = cn(
    'absolute top-1 z-10 inline-flex h-5 w-5 items-center justify-center rounded-full bg-transparent text-slate-400 opacity-0 transition-all group-hover:opacity-100 hover:bg-black/5 hover:text-slate-700 dark:text-slate-500 dark:hover:bg-white/10 dark:hover:text-slate-200',
    isCurrentUser ? '-left-6' : '-right-6',
  );

  if (message.message_type === 'system') {
    return (
      <div className="flex justify-center">
        <span className="text-xs text-slate-400 italic">{message.content}</span>
      </div>
    );
  }

  return (
    <>
      <div className={cn('group flex gap-2', isCurrentUser && 'flex-row-reverse')}>
        {!isCurrentUser && (
          <div
            className={cn(
              'mt-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white',
              getAvatarColorClass(authorName),
            )}
          >
            {avatarInitial}
          </div>
        )}
        <div className={cn('max-w-[75%]', isCurrentUser && 'flex flex-col items-end')}>
          {!isCurrentUser && (
            <div className="mb-0.5 flex items-baseline gap-1.5">
              <span className="text-xs font-medium text-slate-700 dark:text-slate-300">{authorName}</span>
              <span className="text-[10px] text-slate-400">{messageTime}</span>
            </div>
          )}
          <div className="relative inline-block max-w-full">
            <button
              type="button"
              onClick={() => setIsReadsSheetOpen(true)}
              className={detailsButtonClassName}
              aria-label="Message details"
            >
              <MoreVertical className="h-3.5 w-3.5" />
            </button>
            <div className={bubbleClassName}>
              <p className="[overflow-wrap:anywhere] whitespace-pre-line">{message.content}</p>
            </div>
          </div>
          {isCurrentUser && (
            <div className="mt-0.5 flex items-center gap-2 text-[10px] text-slate-400">
              <span>{messageTime}</span>
            </div>
          )}
        </div>
      </div>
      <ChatMessageReadsSheet
        chatId={message.chat_id}
        currentUserId={currentUserId}
        message={message}
        open={isReadsSheetOpen}
        onOpenChange={setIsReadsSheetOpen}
      />
    </>
  );
};

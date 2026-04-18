import type { ChatMessage } from '@/types/chat';
import { cn } from '@/lib/utils';

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

interface ChatMessageItemProps {
  message: ChatMessage;
  isCurrentUser: boolean;
}

export const ChatMessageItem = ({ message, isCurrentUser }: ChatMessageItemProps) => {
  if (message.message_type === 'system') {
    return (
      <div className="flex justify-center">
        <span className="text-sm text-slate-500 italic dark:text-slate-400">{message.content}</span>
      </div>
    );
  }

  return (
    <div className={cn('flex gap-3', isCurrentUser && 'flex-row-reverse')}>
      {!isCurrentUser && (
        <div className="mt-1 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-xs font-medium text-white">
          {message.member?.user?.name.charAt(0).toUpperCase()}
        </div>
      )}
      <div className={cn('max-w-md flex-1', isCurrentUser && 'text-right')}>
        {!isCurrentUser && (
          <div className="mb-1 flex items-center gap-2">
            <span className="text-sm font-medium text-slate-900 dark:text-slate-100">{message.member?.user?.name}</span>
            <span className="text-xs text-slate-500 dark:text-slate-400">{formatTime(message.created_at)}</span>
          </div>
        )}
        <div
          className={cn(
            'rounded-lg px-4 py-2',
            isCurrentUser && 'ml-auto bg-blue-600 text-white',
            !isCurrentUser &&
              'border border-slate-200 bg-white text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100',
          )}
        >
          <p className="text-sm leading-relaxed whitespace-pre-line">{message.content}</p>
        </div>
        {isCurrentUser && (
          <div className="mt-1">
            <span className="text-xs text-slate-500 dark:text-slate-400">{formatTime(message.created_at)}</span>
          </div>
        )}
      </div>
    </div>
  );
};

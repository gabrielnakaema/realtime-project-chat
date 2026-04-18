import { formatTime, getAvatarColorClass } from './utils';
import type { ChatMessage } from '@/types/chat';
import { cn } from '@/lib/utils';

interface MessagesChatMessageItemProps {
  message: ChatMessage;
  isCurrentUser: boolean;
}

export function MessagesChatMessageItem({ message, isCurrentUser }: MessagesChatMessageItemProps) {
  if (message.message_type === 'system') {
    return (
      <div className="flex justify-center">
        <span className="text-xs text-slate-400 italic">{message.content}</span>
      </div>
    );
  }

  return (
    <div className={cn('flex gap-2', isCurrentUser && 'flex-row-reverse')}>
      {!isCurrentUser && (
        <div
          className={cn(
            'mt-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white',
            getAvatarColorClass(message.member?.user?.name ?? ''),
          )}
        >
          {message.member?.user?.name.charAt(0).toUpperCase()}
        </div>
      )}
      <div className={cn('max-w-[75%]', isCurrentUser && 'flex flex-col items-end')}>
        {!isCurrentUser && (
          <div className="mb-0.5 flex items-baseline gap-1.5">
            <span className="text-xs font-medium text-slate-700 dark:text-slate-300">{message.member?.user?.name}</span>
            <span className="text-[10px] text-slate-400">{formatTime(message.created_at)}</span>
          </div>
        )}
        <div
          className={cn(
            'rounded-2xl px-3 py-2 text-sm leading-relaxed',
            isCurrentUser
              ? 'rounded-tr-sm bg-blue-600 text-white'
              : 'rounded-tl-sm border border-slate-200 bg-white text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100',
          )}
        >
          <p className="whitespace-pre-line">{message.content}</p>
        </div>
        {isCurrentUser && <span className="mt-0.5 text-[10px] text-slate-400">{formatTime(message.created_at)}</span>}
      </div>
    </div>
  );
}

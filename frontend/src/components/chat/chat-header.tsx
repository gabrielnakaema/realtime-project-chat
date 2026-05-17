import { ArrowLeft, X } from 'lucide-react';
import { getAvatarColorClass, getChatAvatarSeed, getChatTitle } from './utils';
import type { Chat } from '@/types/chat';
import { cn } from '@/lib/utils';

interface ChatHeaderProps {
  chat?: Chat;
  currentUserId?: string;
  onBack: () => void;
  onClose: () => void;
}

export const ChatHeader = ({ chat, currentUserId, onBack, onClose }: ChatHeaderProps) => {
  const avatarSeed = chat ? getChatAvatarSeed(chat, currentUserId) : '';

  return (
    <div className="flex shrink-0 items-center gap-2 border-b border-slate-100 px-3 py-3 dark:border-slate-700">
      <button
        onClick={onBack}
        className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-800 dark:hover:text-slate-300"
      >
        <ArrowLeft className="h-4 w-4" />
      </button>
      {chat && (
        <>
          <div
            className={cn(
              'flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white',
              getAvatarColorClass(avatarSeed),
            )}
          >
            {avatarSeed.charAt(0).toUpperCase()}
          </div>
          <p className="min-w-0 flex-1 truncate text-sm font-semibold text-slate-900 dark:text-slate-100">
            {getChatTitle(chat, currentUserId)}
          </p>
        </>
      )}
      <button
        onClick={onClose}
        className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-800 dark:hover:text-slate-300"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
};

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
    <div className="border-border flex shrink-0 items-center gap-2 border-b px-3 py-3">
      <button
        onClick={onBack}
        className="text-muted-foreground hover:bg-muted hover:text-muted-foreground flex h-8 w-8 shrink-0 items-center justify-center rounded-lg transition-colors"
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
          <p className="text-foreground min-w-0 flex-1 truncate text-sm font-semibold">
            {getChatTitle(chat, currentUserId)}
          </p>
        </>
      )}
      <button
        onClick={onClose}
        className="text-muted-foreground hover:bg-muted hover:text-muted-foreground flex h-8 w-8 shrink-0 items-center justify-center rounded-lg transition-colors"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
};

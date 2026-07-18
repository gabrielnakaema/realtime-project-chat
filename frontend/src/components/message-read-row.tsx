import { CheckCheck } from 'lucide-react';
import type { ChatMessageRead } from '@/types/chat';

interface MessageReadRowProps {
  read: ChatMessageRead;
  readAtLabel: string;
}

export const MessageReadRow = ({ read, readAtLabel }: MessageReadRowProps) => {
  return (
    <div className="border-border flex items-center justify-between rounded-2xl border px-4 py-3">
      <div className="min-w-0">
        <p className="text-foreground truncate text-sm font-medium">{read.user?.name ?? 'Unknown user'}</p>
        <p className="text-muted-foreground text-xs">{readAtLabel}</p>
      </div>
      <CheckCheck className="text-primary h-4 w-4 shrink-0" />
    </div>
  );
};

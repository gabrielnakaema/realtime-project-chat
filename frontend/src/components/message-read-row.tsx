import { CheckCheck } from 'lucide-react';
import type { ChatMessageRead } from '@/types/chat';

interface MessageReadRowProps {
  read: ChatMessageRead;
  readAtLabel: string;
}

export const MessageReadRow = ({ read, readAtLabel }: MessageReadRowProps) => {
  return (
    <div className="flex items-center justify-between rounded-2xl border border-slate-200 px-4 py-3 dark:border-slate-700">
      <div className="min-w-0">
        <p className="truncate text-sm font-medium text-slate-900 dark:text-slate-100">
          {read.user?.name ?? 'Unknown user'}
        </p>
        <p className="text-xs text-slate-500 dark:text-slate-400">{readAtLabel}</p>
      </div>
      <CheckCheck className="h-4 w-4 shrink-0 text-blue-600 dark:text-blue-400" />
    </div>
  );
};

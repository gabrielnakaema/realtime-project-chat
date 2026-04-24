import { useQuery } from '@tanstack/react-query';
import { LoadingSpinner } from './loading';
import type { ChatMessage } from '@/types/chat';
import { MessageReadRow } from '@/components/message-read-row';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { getMessageReads } from '@/services/chat';
import { chatMessageQueryKeys } from '@/services/query-keys';

interface ChatMessageReadsSheetProps {
  chatId: string;
  currentUserId?: string;
  message: ChatMessage;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const formatReadAt = (timestamp: string) =>
  new Date(timestamp).toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

export const ChatMessageReadsSheet = ({
  chatId,
  currentUserId,
  message,
  open,
  onOpenChange,
}: ChatMessageReadsSheetProps) => {
  const { data: reads = [], isLoading } = useQuery({
    queryKey: chatMessageQueryKeys.reads(chatId, message.id),
    queryFn: () => getMessageReads(chatId, message.id),
    enabled: open && !!chatId,
  });

  const visibleReads = reads.filter((read) => read.user_id !== currentUserId);
  const hasReads = visibleReads.length > 0;

  const renderBody = () => {
    if (isLoading) {
      return (
        <p className="text-sm text-slate-500 dark:text-slate-400">
          <LoadingSpinner size="2rem" />
        </p>
      );
    }

    if (!hasReads) {
      return (
        <div className="rounded-2xl border border-dashed border-slate-200 p-5 text-sm text-slate-500 dark:border-slate-700 dark:text-slate-400">
          No one has read this message yet.
        </div>
      );
    }

    return (
      <div className="space-y-3">
        {visibleReads.map((read) => (
          <MessageReadRow
            key={`${read.message_id}-${read.user_id}`}
            read={read}
            readAtLabel={formatReadAt(read.read_at)}
          />
        ))}
      </div>
    );
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="max-w-[420px]">
        <SheetHeader>
          <SheetTitle>Read status</SheetTitle>
        </SheetHeader>

        <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-6">
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-800">
            <p className="text-sm leading-relaxed [overflow-wrap:anywhere] whitespace-pre-line text-slate-700 dark:text-slate-200">
              {message.content}
            </p>
          </div>

          {renderBody()}
        </div>
      </SheetContent>
    </Sheet>
  );
};

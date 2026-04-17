import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { ArrowLeft, Send, X } from 'lucide-react';
import { Controller, useForm } from 'react-hook-form';
import { formatDate, formatTime, getAvatarColorClass, getChatAvatarSeed, getChatTitle } from './utils';
import type { SubmitHandler } from 'react-hook-form';
import type { IChatForm } from '@/schemas/chat-schema';
import { ExpandingTextarea } from '@/components/expanding-textarea';
import { LoadingSpinner } from '@/components/loading';
import { useMessagesSheet } from '@/contexts/messages-sheet-context';
import { useAuth } from '@/hooks/use-auth';
import { useGeneralChat } from '@/hooks/use-general-chat';
import { cn } from '@/lib/utils';
import { chatSchema } from '@/schemas/chat-schema';
import { createMessage } from '@/services/chat';

interface MessagesChatViewProps {
  chatId: string;
}

export function MessagesChatView({ chatId }: MessagesChatViewProps) {
  const { user } = useAuth();
  const { backToList, close } = useMessagesSheet();

  const { chat, messages, observedRef, chatContainerRef } = useGeneralChat(chatId);

  const { control, handleSubmit, reset } = useForm<IChatForm>({
    resolver: zodResolver(chatSchema),
    defaultValues: { content: '' },
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createMessage,
    onSuccess: () => reset(),
  });

  const onSubmit: SubmitHandler<IChatForm> = (form) => {
    mutate({ chat_id: chatId, content: form.content });
  };

  return (
    <>
      <div className="flex shrink-0 items-center gap-2 border-b border-slate-100 px-3 py-3 dark:border-slate-700">
        <button
          onClick={backToList}
          className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-800 dark:hover:text-slate-300"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        {chat && (
          <>
            <div
              className={cn(
                'flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white',
                getAvatarColorClass(getChatAvatarSeed(chat, user?.id)),
              )}
            >
              {getChatAvatarSeed(chat, user?.id).charAt(0).toUpperCase()}
            </div>
            <p className="min-w-0 flex-1 truncate text-sm font-semibold text-slate-900 dark:text-slate-100">
              {getChatTitle(chat, user?.id)}
            </p>
          </>
        )}
        <button
          onClick={close}
          className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-800 dark:hover:text-slate-300"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div ref={chatContainerRef} className="flex-1 overflow-y-auto p-4">
        <div ref={observedRef} className="h-1" />
        <div className="space-y-3">
          {messages.map((message, index) => {
            const prevDate = index > 0 ? formatDate(messages[index - 1].created_at) : null;
            const msgDate = formatDate(message.created_at);
            const showDate = index === 0 || prevDate !== msgDate;
            const isMe = message.member?.user?.id === user?.id;
            const isSystem = message.message_type === 'system';

            return (
              <div key={message.id}>
                {showDate && (
                  <div className="my-4 flex justify-center">
                    <span className="rounded-full bg-slate-100 px-3 py-0.5 text-xs text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                      {msgDate}
                    </span>
                  </div>
                )}
                {isSystem ? (
                  <div className="flex justify-center">
                    <span className="text-xs text-slate-400 italic">{message.content}</span>
                  </div>
                ) : (
                  <div className={cn('flex gap-2', isMe && 'flex-row-reverse')}>
                    {!isMe && (
                      <div
                        className={cn(
                          'mt-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white',
                          getAvatarColorClass(message.member?.user?.name ?? ''),
                        )}
                      >
                        {message.member?.user?.name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <div className={cn('max-w-[75%]', isMe && 'flex flex-col items-end')}>
                      {!isMe && (
                        <div className="mb-0.5 flex items-baseline gap-1.5">
                          <span className="text-xs font-medium text-slate-700 dark:text-slate-300">
                            {message.member?.user?.name}
                          </span>
                          <span className="text-[10px] text-slate-400">{formatTime(message.created_at)}</span>
                        </div>
                      )}
                      <div
                        className={cn(
                          'rounded-2xl px-3 py-2 text-sm leading-relaxed',
                          isMe
                            ? 'rounded-tr-sm bg-blue-600 text-white'
                            : 'rounded-tl-sm border border-slate-200 bg-white text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100',
                        )}
                      >
                        <p className="whitespace-pre-line">{message.content}</p>
                      </div>
                      {isMe && (
                        <span className="mt-0.5 text-[10px] text-slate-400">{formatTime(message.created_at)}</span>
                      )}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      <div className="shrink-0 border-t border-slate-100 p-3 dark:border-slate-700">
        <form onSubmit={handleSubmit(onSubmit)} className="flex items-end gap-2">
          <Controller
            control={control}
            name="content"
            render={({ field }) => (
              <ExpandingTextarea
                placeholder="Message..."
                wrapperClassName="flex-1"
                className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-400 focus:bg-white focus:ring-1 focus:ring-blue-400 focus:outline-none dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:placeholder-slate-500 dark:focus:bg-slate-700"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    handleSubmit(onSubmit)(e);
                  }
                }}
                {...field}
              />
            )}
          />
          <button
            type="submit"
            disabled={isPending}
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-600 text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
          >
            {isPending ? <LoadingSpinner size="0.875em" /> : <Send className="h-4 w-4" />}
          </button>
        </form>
      </div>
    </>
  );
}

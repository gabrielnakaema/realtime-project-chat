import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Send } from 'lucide-react';
import { Controller, useForm } from 'react-hook-form';
import type { SubmitHandler } from 'react-hook-form';
import type { IChatForm } from '@/schemas/chat-schema';
import { ExpandingTextarea } from '@/components/expanding-textarea';
import { LoadingSpinner } from '@/components/loading';
import { chatSchema } from '@/schemas/chat-schema';
import { createMessage } from '@/services/chat';

interface ChatComposerProps {
  chatId: string;
}

export const ChatComposer = ({ chatId }: ChatComposerProps) => {
  const { control, handleSubmit, reset } = useForm<IChatForm>({
    resolver: zodResolver(chatSchema),
    defaultValues: { content: '' },
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createMessage,
    onSuccess: () => reset(),
  });

  const onSubmit: SubmitHandler<IChatForm> = (form) => {
    if (!chatId) {
      return;
    }

    mutate({ chat_id: chatId, content: form.content });
  };

  return (
    <div className="shrink-0 border-t border-slate-100 p-3 dark:border-slate-700">
      <form onSubmit={handleSubmit(onSubmit)} className="flex items-end gap-2">
        <Controller
          control={control}
          name="content"
          render={({ field }) => (
            <ExpandingTextarea
              placeholder="Message..."
              wrapperClassName="flex-1"
              className="rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm [overflow-wrap:anywhere] text-slate-900 placeholder-slate-400 focus:border-blue-400 focus:bg-white focus:ring-1 focus:ring-blue-400 focus:outline-none dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:placeholder-slate-500 dark:focus:bg-slate-700"
              onKeyDown={(event) => {
                if (event.key === 'Enter' && !event.shiftKey) {
                  event.preventDefault();
                  handleSubmit(onSubmit)(event);
                }
              }}
              {...field}
            />
          )}
        />
        <button
          type="submit"
          disabled={isPending || !chatId}
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-600 text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
        >
          {isPending ? <LoadingSpinner size="0.875em" /> : <Send className="h-4 w-4" />}
        </button>
      </form>
    </div>
  );
};

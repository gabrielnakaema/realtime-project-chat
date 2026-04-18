import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Send } from 'lucide-react';
import { useForm } from 'react-hook-form';
import type { IChatForm } from '@/schemas/chat-schema';
import type { SubmitHandler } from 'react-hook-form';
import { LoadingSpinner } from '@/components/loading';
import { chatSchema } from '@/schemas/chat-schema';
import { createMessage } from '@/services/chat';

interface ChatComposerProps {
  chatId?: string;
}

export const ChatComposer = ({ chatId }: ChatComposerProps) => {
  const { register, handleSubmit, reset } = useForm<IChatForm>({
    resolver: zodResolver(chatSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createMessage,
    onSuccess: () => {
      reset();
    },
  });

  const onSubmit: SubmitHandler<IChatForm> = (form) => {
    if (!chatId) {
      return;
    }

    mutate({
      chat_id: chatId,
      content: form.content,
    });
  };

  return (
    <div className="border-t border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
      <div className="mx-auto max-w-4xl">
        <form onSubmit={handleSubmit(onSubmit)} className="flex items-end gap-3">
          <div className="flex-1">
            <textarea
              placeholder="Type your message..."
              className="w-full rounded-md border border-slate-300 bg-white px-3 py-3 pr-20 text-slate-900 placeholder-slate-500 focus:border-transparent focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 dark:placeholder-slate-400"
              {...register('content')}
              onKeyDown={(event) => {
                if (event.key === 'Enter' && !event.shiftKey) {
                  event.preventDefault();
                  handleSubmit(onSubmit)(event);
                }
              }}
            />
          </div>
          <button
            type="submit"
            className="rounded-md bg-blue-600 px-4 py-3 font-medium text-white transition-colors hover:bg-blue-700 disabled:bg-blue-400"
            disabled={isPending}
          >
            {isPending ? <LoadingSpinner size="1em" /> : <Send className="h-4 w-4" />}
          </button>
        </form>
      </div>
    </div>
  );
};

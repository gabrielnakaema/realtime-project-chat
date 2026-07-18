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
    <div className="border-border shrink-0 border-t p-3">
      <form onSubmit={handleSubmit(onSubmit)} className="flex items-end gap-2">
        <Controller
          control={control}
          name="content"
          render={({ field }) => (
            <ExpandingTextarea
              placeholder="Message..."
              wrapperClassName="flex-1"
              className="border-border bg-muted text-foreground placeholder-muted-foreground focus:border-ring focus:bg-card focus:ring-ring rounded-md border px-3 py-2 text-sm [overflow-wrap:anywhere] focus:ring-1 focus:outline-none"
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
          className="bg-primary text-primary-foreground hover:bg-primary/90 flex h-9 w-9 shrink-0 items-center justify-center rounded-full transition-colors disabled:opacity-50"
        >
          {isPending ? <LoadingSpinner size="0.875em" /> : <Send className="h-4 w-4" />}
        </button>
      </form>
    </div>
  );
};

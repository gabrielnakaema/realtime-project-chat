import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Link, createFileRoute } from '@tanstack/react-router';
import { ArrowLeft, Send, Users } from 'lucide-react';
import { useLayoutEffect, useRef } from 'react';
import { useForm } from 'react-hook-form';
import type { IChatForm } from '@/schemas/chat-schema';
import type { SubmitHandler } from 'react-hook-form';
import { AddProjectMember } from '@/components/add-project-member';
import { LoadingSpinner } from '@/components/loading';
import { MembersAvatarList } from '@/components/members-avatar-list';
import { useAuth } from '@/hooks/use-auth';
import { useChat } from '@/hooks/use-chat';
import { cn } from '@/lib/utils';
import { chatSchema } from '@/schemas/chat-schema';
import { createMessage } from '@/services/chat';

export const Route = createFileRoute('/projects/$projectId/chat')({
  component: RouteComponent,
});

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
};

const formatDate = (timestamp: string) => {
  return new Date(timestamp).toLocaleDateString([], { month: 'short', day: 'numeric' });
};

function RouteComponent() {
  const { projectId } = Route.useParams();
  const { user } = useAuth();
  const isInitialRender = useRef(true);

  const { register, handleSubmit, reset } = useForm<IChatForm>({
    resolver: zodResolver(chatSchema),
  });

  const { project, chatData, messages, observedRef, chatContainerRef, onlineUserIds } = useChat(projectId);

  const { mutate, isPending } = useMutation({
    mutationFn: createMessage,
    onSuccess: () => {
      reset();
    },
  });

  useLayoutEffect(() => {
    const container = chatContainerRef.current;
    if (!container || messages.length === 0) {
      return;
    }

    if (isInitialRender.current) {
      isInitialRender.current = false;
      container.scrollTo({
        top: container.scrollHeight,
        behavior: 'instant',
      });
      return;
    }

    const isNearBottom = container.scrollHeight - container.scrollTop <= container.clientHeight + 100;

    if (isNearBottom) {
      container.scrollTo({
        top: container.scrollHeight,
        behavior: 'smooth',
      });
    }
  }, [messages]);

  const onSubmit: SubmitHandler<IChatForm> = (form) => {
    if (!chatData?.id) {
      return;
    }

    mutate({
      chat_id: chatData.id,
      content: form.content,
    });
  };

  return (
    <div className="flex flex-col h-screen flex-1 ">
      <header className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700  ">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link
                to="/projects/$projectId"
                className="inline-flex items-center px-3 py-2 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-md font-medium transition-colors"
                params={{ projectId }}
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Go back
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">Team chat - {project?.name}</h1>
                <p className="text-slate-600 dark:text-slate-400">Chat with your team members</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <AddProjectMember projectId={projectId} />
                <Users className="w-4 h-4 text-slate-500" />
                <MembersAvatarList
                  onlineUserIds={onlineUserIds}
                  members={
                    project?.members.map((member) => ({
                      user_id: member.user_id,
                      name: member.user.name || '',
                    })) || []
                  }
                  max={4}
                />
              </div>
            </div>
          </div>
        </div>
      </header>

      <div ref={chatContainerRef} className="flex-1 overflow-y-auto p-6 bg-[var(--background)]">
        <div ref={observedRef} className="h-1 bg-transparent" />
        <div className="max-w-4xl mx-auto space-y-4">
          {messages.map((message, index) => {
            const previousMessage = index > 0 ? messages[index - 1] : null;
            const previousMessageDate = previousMessage ? formatDate(previousMessage.created_at) : null;
            const messageDate = formatDate(message.created_at);

            const showDate = index === 0 || previousMessageDate !== messageDate;
            const isCurrentUser = message.member?.user?.id === user?.id;
            const isSystem = message.message_type === 'system';

            return (
              <div key={message.id}>
                {showDate && (
                  <div className="flex justify-center my-6">
                    <span className="bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-400 text-xs px-3 py-1 rounded-full">
                      {messageDate}
                    </span>
                  </div>
                )}

                {isSystem ? (
                  <div className="flex justify-center">
                    <span className="text-sm text-slate-500 dark:text-slate-400 italic">{message.content}</span>
                  </div>
                ) : (
                  <div className={cn('flex gap-3', isCurrentUser && 'flex-row-reverse')}>
                    {!isCurrentUser && (
                      <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-xs font-medium mt-1">
                        {message.member?.user?.name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <div className={cn('flex-1 max-w-md', isCurrentUser && 'text-right')}>
                      {!isCurrentUser && (
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-sm font-medium text-slate-900 dark:text-slate-100">
                            {message.member?.user?.name}
                          </span>
                          <span className="text-xs text-slate-500 dark:text-slate-400">
                            {formatTime(message.created_at)}
                          </span>
                        </div>
                      )}
                      <div
                        className={cn(
                          'rounded-lg px-4 py-2',
                          isCurrentUser && 'bg-blue-600 text-white ml-auto',
                          !isCurrentUser &&
                            'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 border border-slate-200 dark:border-slate-700',
                        )}
                      >
                        <p className="text-sm leading-relaxed">{message.content}</p>
                      </div>
                      {isCurrentUser && (
                        <div className="mt-1">
                          <span className="text-xs text-slate-500 dark:text-slate-400">
                            {formatTime(message.created_at)}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      <div className="border-t border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4">
        <div className="max-w-4xl mx-auto">
          <form onSubmit={handleSubmit(onSubmit)} className="flex items-end gap-3">
            <div className="flex-1">
              <div className="relative">
                <textarea
                  placeholder="Type your message..."
                  className="w-full px-3 py-3 pr-20 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-500 dark:placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  {...register('content')}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) {
                      e.preventDefault();
                      handleSubmit(onSubmit)(e);
                    }
                  }}
                />
              </div>
            </div>
            <button
              type="submit"
              className="px-4 py-3 text-white bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 rounded-md font-medium transition-colors"
              disabled={isPending}
            >
              {isPending ? <LoadingSpinner size="1em" /> : <Send className="w-4 h-4" />}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}

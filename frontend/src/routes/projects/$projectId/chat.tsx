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
import { ProjectMembersModal } from '@/components/project-members-modal';

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
  }, [messages, chatContainerRef]);

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
    <div className="flex h-screen flex-1 flex-col">
      <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link
                to="/projects/$projectId"
                className="inline-flex items-center rounded-md px-3 py-2 font-medium text-slate-700 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
                params={{ projectId }}
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
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
                <ProjectMembersModal project={project} />
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

      <div ref={chatContainerRef} className="flex-1 overflow-y-auto bg-[var(--background)] p-6">
        <div ref={observedRef} className="h-1 bg-transparent" />
        <div className="mx-auto max-w-4xl space-y-4">
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
                  <div className="my-6 flex justify-center">
                    <span className="rounded-full bg-slate-200 px-3 py-1 text-xs text-slate-600 dark:bg-slate-700 dark:text-slate-400">
                      {messageDate}
                    </span>
                  </div>
                )}

                {isSystem ? (
                  <div className="flex justify-center">
                    <span className="text-sm text-slate-500 italic dark:text-slate-400">{message.content}</span>
                  </div>
                ) : (
                  <div className={cn('flex gap-3', isCurrentUser && 'flex-row-reverse')}>
                    {!isCurrentUser && (
                      <div className="mt-1 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-xs font-medium text-white">
                        {message.member?.user?.name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <div className={cn('max-w-md flex-1', isCurrentUser && 'text-right')}>
                      {!isCurrentUser && (
                        <div className="mb-1 flex items-center gap-2">
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
                          isCurrentUser && 'ml-auto bg-blue-600 text-white',
                          !isCurrentUser &&
                            'border border-slate-200 bg-white text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100',
                        )}
                      >
                        <p className="text-sm leading-relaxed whitespace-pre-line">{message.content}</p>
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

      <div className="border-t border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <div className="mx-auto max-w-4xl">
          <form onSubmit={handleSubmit(onSubmit)} className="flex items-end gap-3">
            <div className="flex-1">
              <div className="relative">
                <textarea
                  placeholder="Type your message..."
                  className="w-full rounded-md border border-slate-300 bg-white px-3 py-3 pr-20 text-slate-900 placeholder-slate-500 focus:border-transparent focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 dark:placeholder-slate-400"
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
              className="rounded-md bg-blue-600 px-4 py-3 font-medium text-white transition-colors hover:bg-blue-700 disabled:bg-blue-400"
              disabled={isPending}
            >
              {isPending ? <LoadingSpinner size="1em" /> : <Send className="h-4 w-4" />}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}

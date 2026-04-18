import { createFileRoute } from '@tanstack/react-router';
import { ChatComposer } from '@/components/project-chat/chat-composer';
import { ChatMessageList } from '@/components/project-chat/chat-message-list';
import { ProjectChatHeader } from '@/components/project-chat/project-chat-header';
import { useAuth } from '@/hooks/use-auth';
import { useChatMessages, useChatScrollBehavior, useProjectChat } from '@/hooks/use-chat';

export const Route = createFileRoute('/projects/$projectId/chat')({
  component: RouteComponent,
});

function RouteComponent() {
  const { projectId } = Route.useParams();
  const { user } = useAuth();
  const { project, chat, onlineUserIds } = useProjectChat(projectId);
  const { messages, fetchNextPage } = useChatMessages(projectId, chat?.id);
  const { chatContainerRef, observedRef } = useChatScrollBehavior(messages, fetchNextPage);

  return (
    <div className="flex h-screen flex-1 flex-col">
      <ProjectChatHeader projectId={projectId} project={project} onlineUserIds={onlineUserIds} />
      <ChatMessageList
        messages={messages}
        currentUserId={user?.id}
        chatContainerRef={chatContainerRef}
        observedRef={observedRef}
      />
      <ChatComposer chatId={chat?.id} />
    </div>
  );
}

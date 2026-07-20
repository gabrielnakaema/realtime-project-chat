import { createFileRoute } from '@tanstack/react-router';
import { ChatComposer } from '@/features/chat/components/chat-composer';
import { ChatMessageList } from '@/features/chat/components/chat-message-list';
import { ProjectChatHeader } from '@/features/projects/components/project-chat-header';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useProjectChatView } from '@/features/chat/hooks/use-chat';

export const Route = createFileRoute('/_protected/projects/$projectId/chat')({
  component: RouteComponent,
});

function RouteComponent() {
  const { projectId } = Route.useParams();
  const { user } = useAuth();
  const { project, chat, onlineUserIds, messages, chatContainerRef, observedRef } = useProjectChatView(projectId);

  return (
    <div className="flex h-screen flex-1 flex-col">
      <ProjectChatHeader projectId={projectId} project={project} onlineUserIds={onlineUserIds} />
      <ChatMessageList
        messages={messages}
        currentUserId={user?.id}
        chatContainerRef={chatContainerRef}
        observedRef={observedRef}
      />
      <ChatComposer chatId={chat?.id ?? ''} />
    </div>
  );
}

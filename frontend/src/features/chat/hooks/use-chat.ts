import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent } from 'react';
import { useChatMessageFeed } from './use-chat-message-feed';
import { useChatMessageListBehavior } from './use-chat-message-list-behavior';
import type { SocketEvent } from '@/shared/types/websocket';
import { useOnlineUsers } from '@/shared/hooks/use-online-users';
import { useProjectDetails } from '@/features/projects/hooks/use-project-details';
import { useSocket } from '@/shared/hooks/use-socket';
import { isProjectMembershipEvent } from '@/shared/types/websocket';
import { chatMessageQueryKeys, projectChatQueryKeys, projectQueryKeys } from '@/shared/services/query-keys';
import { getChatByProjectId, listMessagesByProjectId } from '@/features/chat/services/chat';

export const useProjectChatView = (projectId: string) => {
  const queryClient = useQueryClient();

  const { data: project } = useProjectDetails(projectId);

  const { data: chat } = useQuery({
    queryKey: projectChatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const { onlineUserIds } = useOnlineUsers(chat?.id, 'chat');
  const { status, subscribe } = useSocket();
  const handleMembershipEvent = useEffectEvent((event: SocketEvent) => {
    if (!isProjectMembershipEvent(event)) {
      return;
    }

    queryClient.invalidateQueries({ queryKey: projectQueryKeys.details(projectId) });
    queryClient.invalidateQueries({ queryKey: projectChatQueryKeys.detailsByProjectId(projectId) });
  });

  useEffect(() => {
    if (!chat?.id || status !== 'connected') {
      return;
    }

    return subscribe(chat.id, 'chat', handleMembershipEvent);
  }, [chat?.id, status, subscribe]);
  const messagesQueryKey = projectChatQueryKeys.infiniteMessages(projectId);
  const { messages, fetchNextPage } = useChatMessageFeed({
    chatId: chat?.id,
    messagesQueryKey,
    listMessages: (pageParam) => listMessagesByProjectId({ projectId, ...pageParam }),
    onMessageRead: (messageId) => {
      queryClient.invalidateQueries({ queryKey: chatMessageQueryKeys.reads(chat?.id || '', messageId) });
      queryClient.invalidateQueries({
        queryKey: projectChatQueryKeys.detailsByProjectId(projectId),
      });
    },
  });
  const { chatContainerRef, observedRef } = useChatMessageListBehavior({
    chatId: chat?.id,
    messages,
    onLoadMore: fetchNextPage,
  });

  return {
    project,
    chat,
    onlineUserIds,
    messages,
    chatContainerRef,
    observedRef,
  };
};

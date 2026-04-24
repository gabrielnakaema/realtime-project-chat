import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useChatMessageFeed } from './use-chat-message-feed';
import { useChatMessageListBehavior } from './use-chat-message-list-behavior';
import { useOnlineUsers } from './use-online-users';
import { getChatByProjectId, listMessagesByProjectId } from '@/services/chat';
import { getProject } from '@/services/projects';
import { chatMessageQueryKeys, projectChatQueryKeys, projectQueryKeys } from '@/services/query-keys';

export const useProjectChatView = (projectId: string) => {
  const queryClient = useQueryClient();

  const { data: project } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
  });

  const { data: chat } = useQuery({
    queryKey: projectChatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const { onlineUserIds } = useOnlineUsers(chat?.id, 'chat');
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

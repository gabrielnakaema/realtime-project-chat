import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useChatMessageFeed } from './use-chat-message-feed';
import { useChatMessageListBehavior } from './use-chat-message-list-behavior';
import { getGeneralChatById, listGeneralChatMessages } from '@/features/chat/services/general-chat';
import { chatMessageQueryKeys, generalChatQueryKeys } from '@/shared/services/query-keys';

export const useGeneralChat = (chatId: string) => {
  const queryClient = useQueryClient();

  const { data: chat } = useQuery({
    queryKey: generalChatQueryKeys.details(chatId),
    queryFn: () => getGeneralChatById(chatId),
  });

  const { messages, fetchNextPage } = useChatMessageFeed({
    chatId,
    messagesQueryKey: generalChatQueryKeys.infiniteMessages(chatId),
    listMessages: (pageParam) => listGeneralChatMessages({ chatId, ...pageParam }),
    onMessageRead: (messageId) => {
      queryClient.invalidateQueries({
        queryKey: chatMessageQueryKeys.reads(chatId, messageId),
      });
      queryClient.invalidateQueries({
        queryKey: generalChatQueryKeys.list,
        exact: true,
      });
    },
  });
  const { chatContainerRef, observedRef } = useChatMessageListBehavior({
    chatId,
    messages,
    onLoadMore: fetchNextPage,
  });

  return {
    chat,
    observedRef,
    chatContainerRef,
    messages,
  };
};

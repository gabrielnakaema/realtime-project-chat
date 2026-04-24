import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useMemo } from 'react';
import { useSocket } from './use-socket';
import type { ChatMessage } from '@/types/chat';
import type { CursorPaginated } from '@/types/paginated';
import type { SocketEvent } from '@/types/websocket';
import type { InfiniteData, QueryKey } from '@tanstack/react-query';
import { handleError } from '@/utils/handle-error';

interface UseChatMessageFeedProps {
  chatId?: string;
  messagesQueryKey: QueryKey;
  listMessages: (pageParam: { before: string; id: string }) => Promise<CursorPaginated<ChatMessage>>;
  onMessageRead: (messageId: string) => void;
}

export const useChatMessageFeed = ({
  chatId,
  messagesQueryKey,
  listMessages,
  onMessageRead,
}: UseChatMessageFeedProps) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();

  const { data, fetchNextPage } = useInfiniteQuery({
    queryKey: messagesQueryKey,
    queryFn: ({ pageParam }) => listMessages(pageParam),
    getNextPageParam: (lastPage) => {
      if (!lastPage.has_next) {
        return undefined;
      }

      return {
        before: lastPage.data[0].created_at,
        id: lastPage.data[0].id,
      };
    },
    initialPageParam: {
      before: '',
      id: '',
    },
  });

  const addMessage = (message: ChatMessage) => {
    queryClient.setQueryData(messagesQueryKey, (old?: InfiniteData<CursorPaginated<ChatMessage>>) => {
      if (!old?.pages.length) {
        return old;
      }

      const firstPage = old.pages[0];

      return {
        pages: [{ ...firstPage, data: [...firstPage.data, message] }, ...old.pages.slice(1)],
        pageParams: old.pageParams,
      };
    });
  };

  const handleSocketMessage = useEffectEvent((event: SocketEvent) => {
    if (event.type === 'error') {
      handleError(event.data.message);
      return;
    }

    if (event.type === 'message') {
      addMessage(event.data);
      return;
    }

    if (event.type === 'message_read') {
      onMessageRead(event.data.message_id);
    }
  });

  useEffect(() => {
    if (!chatId || status !== 'connected') {
      return;
    }

    const unsubscribe = subscribe(chatId, 'chat', handleSocketMessage);

    return () => {
      unsubscribe();
    };
  }, [chatId, status, subscribe]);

  const messages = useMemo(() => {
    const pages = data?.pages || [];
    const flattenedMessages: ChatMessage[] = [];

    for (let i = pages.length - 1; i >= 0; i--) {
      for (const message of pages[i].data) {
        flattenedMessages.push(message);
      }
    }

    return flattenedMessages;
  }, [data]);

  return {
    messages,
    fetchNextPage,
  };
};

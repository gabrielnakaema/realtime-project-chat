import { useInfiniteQuery, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useLayoutEffect, useMemo, useRef } from 'react';
import { useSocket } from './use-socket';
import type { ChatMessage } from '@/types/chat';
import type { CursorPaginated } from '@/types/paginated';
import type { SocketEvent } from '@/types/websocket';
import type { InfiniteData } from '@tanstack/react-query';
import { getGeneralChatById, listGeneralChatMessages } from '@/services/general-chat';
import { generalChatQueryKeys } from '@/services/query-keys';
import { handleError } from '@/utils/handle-error';

export const useGeneralChat = (chatId: string) => {
  const queryClient = useQueryClient();

  const observedRef = useRef<HTMLDivElement>(null);
  const chatContainerRef = useRef<HTMLDivElement>(null);
  const isInitialRender = useRef(true);
  const wasAtBottomRef = useRef(true);

  const { status, subscribe } = useSocket();

  const { data: chat } = useQuery({
    queryKey: generalChatQueryKeys.details(chatId),
    queryFn: () => getGeneralChatById(chatId),
  });

  const { data: messagesData, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: generalChatQueryKeys.infiniteMessages(chatId),
    queryFn: ({ pageParam }) => listGeneralChatMessages({ chatId, ...pageParam }),
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

  useEffect(() => {
    const container = chatContainerRef.current;
    if (!container) return;

    const handleScroll = () => {
      wasAtBottomRef.current = container.scrollHeight - container.scrollTop - container.clientHeight <= 100;
    };

    container.addEventListener('scroll', handleScroll, { passive: true });
    return () => container.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    const intersectionObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
            fetchNextPage();
          }
        });
      },
      {
        root: chatContainerRef.current,
        rootMargin: '40%',
      },
    );

    if (observedRef.current) {
      intersectionObserver.observe(observedRef.current);
    }

    return () => {
      intersectionObserver.disconnect();
    };
  }, [messagesData, fetchNextPage, hasNextPage, isFetchingNextPage]);

  const addNewMessage = (message: ChatMessage) => {
    queryClient.setQueryData(
      generalChatQueryKeys.infiniteMessages(chatId),
      (old?: InfiniteData<CursorPaginated<ChatMessage>>) => {
        if (!old?.pages.length) {
          return old;
        }

        const firstPage = old.pages[0];
        const hasMorePages = old.pages.length > 1;

        if (hasMorePages) {
          return {
            pages: [{ data: [...firstPage.data, message], has_next: false }, ...old.pages.slice(1)],
            pageParams: old.pageParams,
          };
        }

        return {
          pages: [{ data: [...firstPage.data, message], has_next: false }],
          pageParams: old.pageParams,
        };
      },
    );
  };

  const handleSocketMessage = useEffectEvent((event: SocketEvent) => {
    if (event.type === 'error') {
      handleError(event.data.message);
      return;
    }

    if (event.type === 'message') {
      addNewMessage(event.data);
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
    const pages = messagesData?.pages || [];
    const m: ChatMessage[] = [];
    for (let i = pages.length - 1; i >= 0; i--) {
      for (const message of pages[i].data) {
        m.push(message);
      }
    }
    return m;
  }, [messagesData]);

  useLayoutEffect(() => {
    const container = chatContainerRef.current;
    if (!container || messages.length === 0) {
      return;
    }

    if (isInitialRender.current) {
      isInitialRender.current = false;
      wasAtBottomRef.current = true;
      container.scrollTo({
        top: container.scrollHeight,
        behavior: 'instant',
      });
      return;
    }

    if (wasAtBottomRef.current) {
      container.scrollTo({
        top: container.scrollHeight,
        behavior: 'smooth',
      });
    }
  }, [messages]);

  return {
    chat,
    messagesData,
    observedRef,
    chatContainerRef,
    messages,
  };
};

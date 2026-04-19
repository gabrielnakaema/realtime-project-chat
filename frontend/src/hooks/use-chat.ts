import { useInfiniteQuery, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useLayoutEffect, useMemo, useRef } from 'react';
import { useInfiniteScrollObserver } from './use-infinite-scroll-observer';
import { useOnlineUsers } from './use-online-users';
import { useSocket } from './use-socket';
import type { ChatMessage } from '@/types/chat';
import type { CursorPaginated } from '@/types/paginated';
import type { SocketEvent } from '@/types/websocket';
import type { InfiniteData } from '@tanstack/react-query';
import { getChatByProjectId, listMessagesByProjectId } from '@/services/chat';
import { getProject } from '@/services/projects';
import { chatQueryKeys, projectQueryKeys } from '@/services/query-keys';
import { handleError } from '@/utils/handle-error';

export const useProjectChat = (projectId: string) => {
  const { data: project } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
  });

  const { data: chat } = useQuery({
    queryKey: chatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const { onlineUserIds } = useOnlineUsers(chat?.id, 'chat');

  return {
    project,
    chat,
    onlineUserIds,
  };
};

export const useChatMessages = (projectId: string, chatId?: string) => {
  const queryClient = useQueryClient();
  const { status, subscribe } = useSocket();

  const { data, fetchNextPage } = useInfiniteQuery({
    queryKey: chatQueryKeys.listInfiniteMessagesByProjectId({ projectId }),
    queryFn: ({ pageParam }) => listMessagesByProjectId({ projectId, ...pageParam }),
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
    queryClient.setQueryData(
      chatQueryKeys.listInfiniteMessagesByProjectId({ projectId }),
      (old?: InfiniteData<CursorPaginated<ChatMessage>>) => {
        if (!old?.pages.length) {
          return old;
        }

        const firstPage = old.pages[0];

        if (old.pages.length > 1) {
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
      addMessage(event.data);
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

export const useChatScrollBehavior = (messages: ChatMessage[], onLoadMore: () => void) => {
  const chatContainerRef = useRef<HTMLDivElement>(null);
  const isInitialRender = useRef(true);
  const wasAtBottomRef = useRef(true);

  useEffect(() => {
    const container = chatContainerRef.current;
    if (!container) {
      return;
    }

    const handleScroll = () => {
      wasAtBottomRef.current = container.scrollHeight - container.scrollTop - container.clientHeight <= 100;
    };

    container.addEventListener('scroll', handleScroll, { passive: true });

    return () => {
      container.removeEventListener('scroll', handleScroll);
    };
  }, []);

  const observedRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore,
    rootRef: chatContainerRef,
    rootMargin: '40%',
    threshold: 0,
  });

  useLayoutEffect(() => {
    const container = chatContainerRef.current;
    if (!container || messages.length === 0) {
      return;
    }

    if (isInitialRender.current) {
      isInitialRender.current = false;
      wasAtBottomRef.current = true;
      container.scrollTo({ top: container.scrollHeight, behavior: 'instant' });
      return;
    }

    if (wasAtBottomRef.current) {
      container.scrollTo({ top: container.scrollHeight, behavior: 'smooth' });
    }
  }, [messages]);

  return {
    chatContainerRef,
    observedRef,
  };
};

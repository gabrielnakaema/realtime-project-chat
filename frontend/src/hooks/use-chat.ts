import { useInfiniteQuery, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useEffectEvent, useMemo, useRef } from 'react';
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

export const useChat = (projectId: string) => {
  const queryClient = useQueryClient();

  const observedRef = useRef<HTMLDivElement>(null);
  const chatContainerRef = useRef<HTMLDivElement>(null);

  const { status, subscribe } = useSocket();

  const { data: project } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
  });

  const { data: chatData } = useQuery({
    queryKey: chatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const { data: messagesData, fetchNextPage } = useInfiniteQuery({
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

  const chatId = useMemo(() => chatData?.id, [chatData]);

  const { onlineUserIds } = useOnlineUsers(chatId);

  useEffect(() => {
    const intersectionObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
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
  }, [messagesData, fetchNextPage]);

  const addNewMessage = (message: ChatMessage) => {
    queryClient.setQueryData(
      chatQueryKeys.listInfiniteMessagesByProjectId({ projectId }),
      (old: InfiniteData<CursorPaginated<ChatMessage>>) => {
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
      const chatMessage = event.data;

      addNewMessage(chatMessage);
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

  return {
    project,
    chatData,
    messagesData,
    observedRef,
    chatContainerRef,
    messages,
    onlineUserIds,
  };
};

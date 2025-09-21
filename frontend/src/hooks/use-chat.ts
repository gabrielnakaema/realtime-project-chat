import { tokenService } from '@/services/api';
import { getChatByProjectId, listMessagesByProjectId } from '@/services/chat';
import { getProject } from '@/services/projects';
import { chatQueryKeys, projectQueryKeys } from '@/services/query-keys';
import type { ChatMessage, SocketEvent } from '@/types/chat';
import type { CursorPaginated } from '@/types/paginated';
import { handleError } from '@/utils/handle-error';
import { useInfiniteQuery, useQuery, useQueryClient, type InfiniteData } from '@tanstack/react-query';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useAuth } from './use-auth';

export const useChat = (projectId: string) => {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuth();

  const [before, setBefore] = useState<string | null>(null);
  const observedRef = useRef<HTMLDivElement>(null);
  const chatContainerRef = useRef<HTMLDivElement>(null);
  const socket = useRef<WebSocket>(null);

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
      if (lastPage && !lastPage.has_next) {
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

  useEffect(() => {
    if (!isAuthenticated || !chatId) {
      return;
    }

    const s = socket.current;

    if (s && (s.readyState === WebSocket.OPEN || s.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const newSocket = new WebSocket(`${import.meta.env.VITE_API_URL}/ws?chat_id=${chatId}&jwt=${tokenService.token}`);

    newSocket.onmessage = (event: MessageEvent) => {
      const data = JSON.parse(event.data) as SocketEvent;

      if (data.type === 'error') {
        handleError(data.data.message);
        return;
      }

      if (data.type === 'message') {
        const chatMessage = data.data;
        if (!chatMessage) {
          return;
        }

        handleNewMessage(chatMessage);
      }
    };

    socket.current = newSocket;

    return () => {
      newSocket.close();
      socket.current = null;
    };
  }, [chatId, isAuthenticated]);

  const handleNewMessage = (message: ChatMessage) => {
    queryClient.setQueryData(
      chatQueryKeys.listInfiniteMessagesByProjectId({ projectId }),
      (old: InfiniteData<CursorPaginated<ChatMessage>>) => {
        const firstPage = old?.pages?.[0];

        if (!firstPage) {
          return {
            pages: [{ data: [message], has_next: false }],
            pageParams: old.pageParams,
          };
        }

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

  const messages = useMemo(() => {
    const pages = messagesData?.pages || [];

    const m: ChatMessage[] = [];

    for (let i = pages.length - 1; i >= 0; i--) {
      for (const message of pages[i].data || []) {
        m.push(message);
      }
    }

    return m;
  }, [messagesData]);

  return {
    project,
    chatData,
    messagesData,
    before,
    setBefore,
    observedRef,
    chatContainerRef,
    messages,
  };
};

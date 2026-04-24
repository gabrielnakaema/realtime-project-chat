import { useEffect, useLayoutEffect, useRef } from 'react';
import { useChatReadTracking } from './use-chat-read-tracking';
import { useInfiniteScrollObserver } from './use-infinite-scroll-observer';
import type { ChatMessage } from '@/types/chat';

interface UseChatMessageListBehaviorProps {
  chatId?: string;
  messages: ChatMessage[];
  onLoadMore: () => void;
}

export const useChatMessageListBehavior = ({ chatId, messages, onLoadMore }: UseChatMessageListBehaviorProps) => {
  const chatContainerRef = useRef<HTMLDivElement>(null);
  const isInitialRender = useRef(true);
  const isAtBottomRef = useRef(true);

  const { maybeMarkLatestAsRead } = useChatReadTracking({
    chatId,
    latestMessageId: messages.at(-1)?.id,
    isAtBottomRef,
  });

  useEffect(() => {
    const container = chatContainerRef.current;
    if (!container) {
      return;
    }

    const handleScroll = () => {
      isAtBottomRef.current = container.scrollHeight - container.scrollTop - container.clientHeight <= 100;

      if (isAtBottomRef.current) {
        maybeMarkLatestAsRead();
      }
    };

    container.addEventListener('scroll', handleScroll, { passive: true });

    return () => {
      container.removeEventListener('scroll', handleScroll);
    };
  }, [maybeMarkLatestAsRead]);

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
      isAtBottomRef.current = true;
      container.scrollTo({
        top: container.scrollHeight,
        behavior: 'instant',
      });
      return;
    }

    if (isAtBottomRef.current) {
      container.scrollTo({
        top: container.scrollHeight,
        behavior: 'smooth',
      });
    }
  }, [messages]);

  return {
    chatContainerRef,
    observedRef,
  };
};

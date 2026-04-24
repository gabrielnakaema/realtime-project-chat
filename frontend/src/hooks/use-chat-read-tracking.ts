import { useMutation } from '@tanstack/react-query';
import { useCallback, useEffect, useRef } from 'react';
import type { RefObject } from 'react';
import { markChatRead } from '@/services/chat';

interface UseChatReadTrackingProps {
  chatId?: string;
  latestMessageId?: string;
  isAtBottomRef: RefObject<boolean>;
}

export const useChatReadTracking = ({ chatId, latestMessageId, isAtBottomRef }: UseChatReadTrackingProps) => {
  const lastSubmittedMessageIdRef = useRef<string | null>(null);

  const { mutate } = useMutation({
    mutationFn: ({ currentChatId, messageId }: { currentChatId: string; messageId: string }) =>
      markChatRead(currentChatId, messageId),
    onError: () => {
      lastSubmittedMessageIdRef.current = null;
    },
  });

  const maybeMarkLatestAsRead = useCallback(() => {
    if (!chatId || !latestMessageId || !isAtBottomRef.current) {
      return;
    }

    if (lastSubmittedMessageIdRef.current === latestMessageId) {
      return;
    }

    lastSubmittedMessageIdRef.current = latestMessageId;
    mutate({ currentChatId: chatId, messageId: latestMessageId });
  }, [chatId, latestMessageId, isAtBottomRef, mutate]);

  useEffect(() => {
    lastSubmittedMessageIdRef.current = null;
  }, [chatId]);

  useEffect(() => {
    maybeMarkLatestAsRead();
  }, [maybeMarkLatestAsRead]);

  useEffect(() => {
    const onFocus = () => {
      maybeMarkLatestAsRead();
    };

    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [maybeMarkLatestAsRead]);

  return {
    maybeMarkLatestAsRead,
  };
};

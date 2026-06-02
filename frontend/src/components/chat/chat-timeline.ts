import type { ChatMessage } from '@/types/chat';
import { formatDate } from './utils';

type MessagesTimelineItem =
  | {
      id: string;
      type: 'date';
      label: string;
    }
  | {
      id: string;
      type: 'message';
      message: ChatMessage;
      isCurrentUser: boolean;
    };

export const buildMessagesTimeline = (messages: ChatMessage[], currentUserId?: string) => {
  const timeline: MessagesTimelineItem[] = [];

  for (let index = 0; index < messages.length; index++) {
    const message = messages[index];
    const previousMessage = index > 0 ? messages[index - 1] : null;
    const previousDate = previousMessage ? formatDate(previousMessage.created_at) : null;
    const currentDate = formatDate(message.created_at);

    if (!previousMessage || previousDate !== currentDate) {
      timeline.push({
        id: `date-${message.id}`,
        type: 'date',
        label: currentDate,
      });
    }

    timeline.push({
      id: message.id,
      type: 'message',
      message,
      isCurrentUser: message.member?.user?.id === currentUserId,
    });
  }

  return timeline;
};

import type { ChatMessage } from '@/types/chat';

const formatDate = (timestamp: string) => {
  return new Date(timestamp).toLocaleDateString([], { month: 'short', day: 'numeric' });
};

type ChatTimelineItem =
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

export const buildChatTimeline = (messages: ChatMessage[], currentUserId?: string) => {
  const timeline: ChatTimelineItem[] = [];

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

import type { Chat, ChatMessage, ChatMessageRead } from '@/features/chat/types/chat';
import type { CursorPaginated } from '@/shared/types/paginated';
import { api } from '@/shared/services/api';

export const getChatByProjectId = async (projectId: string) => {
  const response = await api.get(`projects/${projectId}/chat`);

  const json = await response.json<Chat>();

  return json;
};

interface ListMessagesByProjectIdPayload {
  before: string;
  id: string;
  projectId: string;
}

export const listMessagesByProjectId = async (payload: ListMessagesByProjectIdPayload) => {
  const searchParams = new URLSearchParams();
  if (payload.before) {
    searchParams.set('before', payload.before);
  }

  if (payload.id) {
    searchParams.set('id', payload.id);
  }

  searchParams.set('limit', '30');

  const response = await api.get(`projects/${payload.projectId}/chat/messages`, {
    searchParams,
  });

  const json = await response.json<CursorPaginated<ChatMessage>>();

  return json;
};

interface CreateMessagePayload {
  chat_id: string;
  content: string;
}

export const createMessage = async (payload: CreateMessagePayload) => {
  const response = await api.post(`chats/messages`, {
    json: payload,
  });

  const json = await response.json<ChatMessage>();

  return json;
};

export const markChatRead = async (chatId: string, messageId?: string) => {
  await api.post(`chats/${chatId}/read`, {
    json: {
      message_id: messageId ?? null,
    },
  });
};

export const getMessageReads = async (chatId: string, messageId: string) => {
  const response = await api.get(`chats/${chatId}/messages/${messageId}/reads`);

  return response.json<ChatMessageRead[]>();
};

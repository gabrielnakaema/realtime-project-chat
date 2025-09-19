import type { Chat, ChatMessage } from '@/types/chat';
import { api } from './api';
import type { CursorPaginated } from '@/types/paginated';

export const getChatByProjectId = async (projectId: string) => {
  const response = await api.get(`projects/${projectId}/chat`);

  const json = await response.json<Chat>();

  return json;
};

export const listMessagesByProjectId = async (projectId: string, before: string) => {
  const searchParams = new URLSearchParams();
  if (before) {
    searchParams.set('before', before);
  }

  const response = await api.get(`projects/${projectId}/chat/messages`, {
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

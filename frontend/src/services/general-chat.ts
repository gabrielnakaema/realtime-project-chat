import { api } from './api';
import type { Chat, ChatMessage } from '@/types/chat';
import type { CursorPaginated } from '@/types/paginated';

export const getOrCreateGeneralChat = async (userIds: string[]): Promise<Chat> => {
  const response = await api.post('chats', {
    json: { user_ids: userIds },
  });
  return response.json<Chat>();
};

export const listGeneralChats = async (): Promise<Chat[]> => {
  const response = await api.get('chats');
  return response.json<Chat[]>();
};

export const getGeneralChatById = async (chatId: string): Promise<Chat> => {
  const response = await api.get(`chats/${chatId}`);
  return response.json<Chat>();
};

interface ListGeneralChatMessagesPayload {
  chatId: string;
  before: string;
  id: string;
}

export const listGeneralChatMessages = async (
  payload: ListGeneralChatMessagesPayload,
): Promise<CursorPaginated<ChatMessage>> => {
  const searchParams = new URLSearchParams();
  if (payload.before) {
    searchParams.set('before', payload.before);
  }
  if (payload.id) {
    searchParams.set('id', payload.id);
  }
  searchParams.set('limit', '30');

  const response = await api.get(`chats/${payload.chatId}/messages`, { searchParams });
  return response.json<CursorPaginated<ChatMessage>>();
};

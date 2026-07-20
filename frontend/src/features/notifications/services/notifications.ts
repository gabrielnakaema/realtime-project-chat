import type { CursorPaginated } from '@/shared/types/paginated';
import type { Notification } from '@/features/notifications/types/notification';
import { api } from '@/shared/services/api';

interface ListNotificationsPayload {
  before?: string;
  id?: string;
  limit?: number;
}

export const listNotifications = async (payload: ListNotificationsPayload) => {
  const searchParams = new URLSearchParams();

  if (payload.before) {
    searchParams.set('before', payload.before);
  }

  if (payload.id) {
    searchParams.set('id', payload.id);
  }

  searchParams.set('limit', payload.limit?.toString() ?? '10');

  const response = await api.get('notifications', {
    searchParams,
  });

  return response.json<CursorPaginated<Notification>>();
};

export const getUnreadNotificationCount = async () => {
  const response = await api.get('notifications/unread-count');
  return response.json<{ count: number }>();
};

export const markNotificationRead = async (notificationId: string) => {
  const response = await api.post(`notifications/${notificationId}/read`);
  return response.json<{ ok: boolean }>();
};

export const markAllNotificationsRead = async () => {
  const response = await api.post('notifications/read-all');
  return response.json<{ ok: boolean }>();
};

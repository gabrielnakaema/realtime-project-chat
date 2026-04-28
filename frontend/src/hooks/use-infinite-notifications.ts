import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo } from 'react';
import type { InfiniteData } from '@tanstack/react-query';
import type { Notification } from '@/types/notification';
import type { CursorPaginated } from '@/types/paginated';
import { listNotifications } from '@/services/notifications';
import { notificationQueryKeys } from '@/services/query-keys';

const PAGE_SIZE = 10;

export const useInfiniteNotifications = (enabled: boolean) => {
  const queryClient = useQueryClient();

  const notificationListQuery = useInfiniteQuery({
    queryKey: notificationQueryKeys.list,
    initialPageParam: {
      before: new Date().toISOString(),
      id: '',
    },
    enabled,
    queryFn: ({ pageParam }) => listNotifications({ before: pageParam.before, id: pageParam.id, limit: PAGE_SIZE }),
    getNextPageParam: (lastPage) => {
      if (!lastPage.has_next || !lastPage.data.length) {
        return undefined;
      }

      const lastNotification = lastPage.data[lastPage.data.length - 1];

      return {
        before: lastNotification.created_at,
        id: lastNotification.id,
      };
    },
  });

  const notifications = useMemo(() => {
    return notificationListQuery.data?.pages.flatMap((page) => page.data) ?? [];
  }, [notificationListQuery.data]);

  const prependNotification = (notification: Notification) => {
    queryClient.setQueryData<InfiniteData<CursorPaginated<Notification>> | undefined>(
      notificationQueryKeys.list,
      (current) => {
        if (!current) {
          return current;
        }

        const firstPage = current.pages[0];
        if (firstPage.data.some((item) => item.id === notification.id)) {
          return current;
        }

        return {
          ...current,
          pages: [
            {
              ...firstPage,
              data: [notification, ...firstPage.data],
            },
            ...current.pages.slice(1),
          ],
        };
      },
    );

    queryClient.setQueryData<{ count: number } | undefined>(notificationQueryKeys.unreadCount, (current) => ({
      count: (current?.count ?? 0) + 1,
    }));
  };

  const markNotificationAsRead = (notification: Notification) => {
    const readAt = notification.read_at ?? new Date().toISOString();

    queryClient.setQueryData<InfiniteData<CursorPaginated<Notification>> | undefined>(
      notificationQueryKeys.list,
      (current) => {
        if (!current) {
          return current;
        }

        return {
          ...current,
          pages: current.pages.map((page) => ({
            ...page,
            data: page.data.map((item) => (item.id === notification.id ? { ...item, read_at: readAt } : item)),
          })),
        };
      },
    );

    queryClient.setQueryData<{ count: number } | undefined>(notificationQueryKeys.unreadCount, (current) => ({
      count: Math.max((current?.count ?? 1) - 1, 0),
    }));
  };

  const markAllNotificationsAsRead = () => {
    queryClient.setQueryData<InfiniteData<CursorPaginated<Notification>> | undefined>(
      notificationQueryKeys.list,
      (current) => {
        if (!current) {
          return current;
        }

        const readAt = new Date().toISOString();

        return {
          ...current,
          pages: current.pages.map((page) => ({
            ...page,
            data: page.data.map((notification) => ({ ...notification, read_at: notification.read_at ?? readAt })),
          })),
        };
      },
    );

    queryClient.setQueryData(notificationQueryKeys.unreadCount, { count: 0 });
  };

  return {
    notifications,
    hasNextPage: notificationListQuery.hasNextPage,
    isLoading: notificationListQuery.isLoading,
    isFetchingNextPage: notificationListQuery.isFetchingNextPage,
    fetchNextPage: notificationListQuery.fetchNextPage,
    prependNotification,
    markNotificationAsRead,
    markAllNotificationsAsRead,
  };
};

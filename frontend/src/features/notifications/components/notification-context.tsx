import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { createContext, useEffect, useEffectEvent, useRef, useState } from 'react';
import type { Notification } from '@/features/notifications/types/notification';
import type { SocketEvent } from '@/shared/types/websocket';
import { getNotificationNavigationTarget } from '@/features/notifications/components/notification-navigation';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useInfiniteNotifications } from '@/features/notifications/hooks/use-infinite-notifications';
import { useSocket } from '@/shared/hooks/use-socket';
import {
  getUnreadNotificationCount,
  markAllNotificationsRead,
  markNotificationRead,
} from '@/features/notifications/services/notifications';
import { notificationQueryKeys } from '@/shared/services/query-keys';

interface NotificationContextData {
  isOpen: boolean;
  setIsOpen: (open: boolean) => void;
  notifications: Notification[];
  unreadCount: number;
  hasNextPage: boolean;
  isLoading: boolean;
  isFetchingNextPage: boolean;
  fetchNextPage: () => Promise<unknown>;
  openNotification: (notification: Notification) => Promise<void>;
  markAllAsRead: () => Promise<void>;
}

export const NotificationContext = createContext<NotificationContextData>({} as NotificationContextData);

export const NotificationProvider = ({ children }: { children: React.ReactNode }) => {
  const { user, isAuthenticated } = useAuth();
  const { status, subscribe } = useSocket();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);
  const connectedOnce = useRef(false);
  const notificationList = useInfiniteNotifications(isAuthenticated);

  const unreadCountQuery = useQuery({
    queryKey: notificationQueryKeys.unreadCount,
    enabled: isAuthenticated,
    queryFn: getUnreadNotificationCount,
  });

  const markReadMutation = useMutation({
    mutationFn: markNotificationRead,
  });

  const markAllReadMutation = useMutation({
    mutationFn: markAllNotificationsRead,
  });

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.type !== 'notification_created') {
      return;
    }

    notificationList.prependNotification(event.data);
  });

  useEffect(() => {
    if (!user?.id || status !== 'connected') {
      return;
    }

    const unsubscribe = subscribe(user.id, 'user', handleSocketEvent);
    return () => unsubscribe();
  }, [status, subscribe, user?.id]);

  useEffect(() => {
    if (status !== 'connected' || !isAuthenticated) {
      return;
    }

    if (connectedOnce.current) {
      queryClient.invalidateQueries({ queryKey: notificationQueryKeys.list });
      queryClient.invalidateQueries({ queryKey: notificationQueryKeys.unreadCount });
    }

    connectedOnce.current = true;
  }, [isAuthenticated, queryClient, status]);

  const markNotificationReadAndUpdateCache = async (notification: Notification) => {
    if (notification.read_at) {
      return;
    }

    await markReadMutation.mutateAsync(notification.id);
    notificationList.markNotificationAsRead(notification);
  };

  const openNotification = async (notification: Notification) => {
    await markNotificationReadAndUpdateCache(notification);
    setIsOpen(false);

    await navigate(getNotificationNavigationTarget(notification));
  };

  const markAllAsRead = async () => {
    await markAllReadMutation.mutateAsync();

    notificationList.markAllNotificationsAsRead();
  };

  return (
    <NotificationContext.Provider
      value={{
        isOpen,
        setIsOpen,
        notifications: notificationList.notifications,
        unreadCount: unreadCountQuery.data?.count ?? 0,
        hasNextPage: notificationList.hasNextPage,
        isLoading: notificationList.isLoading || unreadCountQuery.isLoading,
        isFetchingNextPage: notificationList.isFetchingNextPage,
        fetchNextPage: notificationList.fetchNextPage,
        openNotification,
        markAllAsRead,
      }}
    >
      {children}
    </NotificationContext.Provider>
  );
};

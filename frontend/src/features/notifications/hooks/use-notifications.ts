import { useContext } from 'react';
import { NotificationContext } from '@/features/notifications/components/notification-context';

export const useNotifications = () => {
  return useContext(NotificationContext);
};

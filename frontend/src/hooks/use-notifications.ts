import { useContext } from 'react';
import { NotificationContext } from '@/contexts/notification-context';

export const useNotifications = () => {
  return useContext(NotificationContext);
};

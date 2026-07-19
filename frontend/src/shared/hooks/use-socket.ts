import { useContext } from 'react';
import { SocketContext } from '@/shared/contexts/socket-context';

export const useSocket = () => {
  const context = useContext(SocketContext);

  return context;
};

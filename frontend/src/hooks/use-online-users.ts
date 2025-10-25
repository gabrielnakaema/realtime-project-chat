import { useEffect, useEffectEvent, useState } from 'react';
import { useSocket } from './use-socket';
import type { SocketEvent } from '@/types/websocket';

export const useOnlineUsers = (roomId?: string) => {
  const [onlineUserIds, setOnlineUserIds] = useState<string[]>([]);

  const { status, subscribe } = useSocket();

  const handleSocketEvent = useEffectEvent((event: SocketEvent) => {
    if (event.room_id !== roomId) {
      return;
    }

    if (event.type === 'users_online') {
      setOnlineUserIds(Array.isArray(event.data) ? event.data : []);
    }

    if (event.type === 'user_connected') {
      setOnlineUserIds((prev) => {
        if (prev.includes(event.data.user_id)) {
          return prev;
        }

        return [...prev, event.data.user_id];
      });
    }

    if (event.type === 'user_disconnected') {
      setOnlineUserIds((prev) => prev.filter((u) => u !== event.data.user_id));
    }
  });

  useEffect(() => {
    if (status !== 'connected' || !roomId) {
      return;
    }

    const unsubscribe = subscribe(roomId, '', handleSocketEvent);

    return () => {
      unsubscribe();
    };
  }, [status, roomId, subscribe]);

  return {
    onlineUserIds,
  };
};

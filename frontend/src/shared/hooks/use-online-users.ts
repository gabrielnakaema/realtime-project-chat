import { useEffect, useEffectEvent, useState } from 'react';
import type { SocketEvent } from '@/shared/types/websocket';
import { useSocket } from '@/shared/hooks/use-socket';

type OnlineUsersRoomType = 'chat' | 'project';

export const useOnlineUsers = (roomId?: string, roomType?: OnlineUsersRoomType) => {
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
    if (status !== 'connected' || !roomId || !roomType) {
      return;
    }

    const unsubscribe = subscribe(roomId, roomType, handleSocketEvent);

    return () => {
      unsubscribe();
    };
  }, [status, roomId, roomType, subscribe]);

  return {
    onlineUserIds,
  };
};

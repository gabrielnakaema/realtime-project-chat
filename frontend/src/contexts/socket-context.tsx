import { createContext, useCallback, useEffect, useEffectEvent, useRef, useState } from 'react';
import type { SocketEvent } from '@/types/websocket';
import { useAuth } from '@/hooks/use-auth';
import { tokenService } from '@/services/api';

type WebSocketStatus = 'disconnected' | 'connected';

type SocketRoomType = 'chat' | 'project' | '';

type SocketHandler = (event: SocketEvent) => void;

interface SocketPayload<T> {
  type: string;
  room_id?: string;
  data: T;
}

interface Subscription {
  roomId: string;
  type: SocketRoomType;
  handler: SocketHandler;
  id: string;
}

interface SocketContextData {
  status: WebSocketStatus;
  subscribe: (roomId: string, type: SocketRoomType, handler: SocketHandler) => () => void;
}

export const SocketContext = createContext<SocketContextData>({} as SocketContextData);

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuth();
  const socket = useRef<WebSocket>(null);

  const [status, setStatus] = useState<WebSocketStatus>('disconnected');
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);

  const handleOpen = useEffectEvent(() => {
    socket.current?.send(JSON.stringify({ type: 'ping', data: null }));
    setStatus('connected');
  });

  const handleClose = useEffectEvent(() => {
    setStatus('disconnected');
  });

  const handleError = useEffectEvent(() => {
    setStatus('disconnected');
  });

  const handleMessage = useEffectEvent((event: MessageEvent) => {
    if (socket.current?.readyState !== WebSocket.OPEN) {
      return;
    }

    try {
      const data = JSON.parse(event.data) as SocketEvent;

      if (data.type === 'ping') {
        socket.current.send(JSON.stringify({ type: 'pong', data: null }));
        return;
      }

      subscriptions.forEach((subscription) => subscription.handler(data));
    } catch (error) {
      return;
    }
  });

  useEffect(() => {
    if (!isAuthenticated) {
      socket.current?.close();
      return;
    }

    const newSocket = new WebSocket(`${import.meta.env.VITE_API_URL}/ws?jwt=${tokenService.token}`);

    newSocket.onopen = handleOpen;
    newSocket.onclose = handleClose;
    newSocket.onerror = handleError;
    newSocket.onmessage = handleMessage;

    socket.current = newSocket;

    return () => {
      socket.current?.close();
      socket.current = null;
    };
  }, [isAuthenticated]);

  const send = useCallback(
    (payload: SocketPayload<any>) => {
      if (status === 'connected' && socket.current) {
        socket.current.send(JSON.stringify(payload));
      }
    },
    [status],
  );

  const subscribe = useCallback(
    (roomId: string, type: SocketRoomType, handler: SocketHandler) => {
      const subscriptionId = crypto.randomUUID();

      const newSubscription: Subscription = {
        id: subscriptionId,
        roomId,
        type,
        handler,
      };

      setSubscriptions((prev) => {
        const existingSubToRoom = prev.find((sub) => sub.roomId === roomId);

        if (!existingSubToRoom) {
          send({ type: 'connect_user_to_room', data: { room_id: roomId, type } });
        }

        return [...prev, newSubscription];
      });

      return () => {
        setSubscriptions((prev) => {
          const updated = prev.filter((sub) => sub.id !== subscriptionId);
          const remaining = updated.filter((sub) => sub.roomId !== roomId);

          if (!remaining.length) {
            send({ type: 'disconnect_user_from_room', data: { room_id: roomId } });
          }

          return updated;
        });
      };
    },
    [send],
  );

  return <SocketContext.Provider value={{ status, subscribe }}>{children}</SocketContext.Provider>;
};

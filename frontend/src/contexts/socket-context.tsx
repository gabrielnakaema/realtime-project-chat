import { createContext, useCallback, useEffect, useEffectEvent, useRef, useState } from 'react';
import type { SocketEvent } from '@/types/websocket';
import { useAuth } from '@/hooks/use-auth';
import { tokenService } from '@/services/api';

type WebSocketStatus = 'disconnected' | 'connected';

type SocketRoomType = 'chat' | 'project' | 'user' | '';

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

const RECONNECT_BASE_DELAY_MS = 1_000;
const RECONNECT_MAX_DELAY_MS = 30_000;

export const SocketContext = createContext<SocketContextData>({} as SocketContextData);

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuth();
  const socket = useRef<WebSocket>(null);

  const [status, setStatus] = useState<WebSocketStatus>('disconnected');

  const [connectionAttempt, setConnectionAttempt] = useState(0);
  const subscriptions = useRef<Map<string, Subscription>>(new Map());
  const reconnectAttempt = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const send = useCallback((payload: SocketPayload<any>) => {
    if (socket.current?.readyState === WebSocket.OPEN) {
      socket.current.send(JSON.stringify(payload));
    }
  }, []);

  const handleOpen = useEffectEvent(() => {
    reconnectAttempt.current = 0;
    socket.current?.send(JSON.stringify({ type: 'ping', data: null }));
    const rooms = new Map<string, SocketRoomType>();
    for (const subscription of subscriptions.current.values()) {
      if (!rooms.has(subscription.roomId)) {
        rooms.set(subscription.roomId, subscription.type);
      }
    }
    for (const [roomId, type] of rooms) {
      send({ type: 'connect_user_to_room', data: { room_id: roomId, type } });
    }

    setStatus('connected');
  });

  const handleClose = useEffectEvent(() => {
    setStatus('disconnected');

    if (reconnectTimer.current) {
      return;
    }

    const delay = Math.min(RECONNECT_BASE_DELAY_MS * 2 ** reconnectAttempt.current, RECONNECT_MAX_DELAY_MS);
    reconnectAttempt.current += 1;
    reconnectTimer.current = setTimeout(() => {
      reconnectTimer.current = null;
      setConnectionAttempt((attempt) => attempt + 1);
    }, delay);
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

      for (const subscription of subscriptions.current.values()) {
        if (subscription.roomId !== data.room_id) {
          continue;
        }

        subscription.handler(data);
      }
    } catch (error) {
      return;
    }
  });

  useEffect(() => {
    if (!isAuthenticated) {
      return;
    }

    const newSocket = new WebSocket(`${import.meta.env.VITE_API_URL}/ws?jwt=${tokenService.token}`);

    newSocket.onopen = handleOpen;
    newSocket.onclose = handleClose;
    newSocket.onerror = handleError;
    newSocket.onmessage = handleMessage;

    socket.current = newSocket;

    return () => {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
      newSocket.onopen = null;
      newSocket.onclose = null;
      newSocket.onerror = null;
      newSocket.onmessage = null;
      newSocket.close();

      if (socket.current === newSocket) {
        socket.current = null;
      }
      setStatus('disconnected');
    };
  }, [isAuthenticated, connectionAttempt]);

  const subscribe = useCallback(
    (roomId: string, type: SocketRoomType, handler: SocketHandler) => {
      const subscriptionId = crypto.randomUUID();
      const existing = [...subscriptions.current.values()];
      const hasMatchingSubscription = existing.some((sub) => sub.roomId === roomId && sub.type === type);
      const existingSubToRoom = existing.find((sub) => sub.roomId === roomId);
      const shouldJoinRoom =
        !hasMatchingSubscription && (!existingSubToRoom || (!!type && existingSubToRoom.type !== type));

      subscriptions.current.set(subscriptionId, { id: subscriptionId, roomId, type, handler });

      if (shouldJoinRoom) {
        send({ type: 'connect_user_to_room', data: { room_id: roomId, type } });
      }

      return () => {
        subscriptions.current.delete(subscriptionId);

        const roomStillInUse = [...subscriptions.current.values()].some((sub) => sub.roomId === roomId);
        if (!roomStillInUse) {
          send({ type: 'disconnect_user_from_room', data: { room_id: roomId } });
        }
      };
    },
    [send],
  );

  return <SocketContext.Provider value={{ status, subscribe }}>{children}</SocketContext.Provider>;
};

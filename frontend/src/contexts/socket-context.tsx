import { useAuth } from '@/hooks/use-auth';
import { tokenService } from '@/services/api';
import type { SocketEvent } from '@/types/websocket';
import { createContext, useCallback, useEffect, useRef, useState } from 'react';

type WebSocketStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

type SocketRoomType = 'chat' | 'project';

interface SocketRoom {
  id: string;
  type: SocketRoomType;
}

interface HandlerItem {
  id: string;
  handler: SocketHandler;
}

type SocketHandler = (event: SocketEvent) => void;

interface SocketPayload<T> {
  type: string;
  room_id?: string;
  data: T;
}

interface SocketContextData {
  status: WebSocketStatus;
  socket: WebSocket | null;
  send: (payload: SocketPayload<any>) => void;
  registerHandler: (handler: HandlerItem) => void;
  unregisterHandlers: (ids: string[]) => void;
  connectToRoom: (roomId: string, type: SocketRoomType) => void;
  disconnectFromRoom: (roomId: string) => void;
}

export const SocketContext = createContext<SocketContextData>({} as SocketContextData);

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuth();
  const [status, setStatus] = useState<WebSocketStatus>('disconnected');
  const socket = useRef<WebSocket>(null);
  const [handlers, setHandlers] = useState<HandlerItem[]>([]);
  const [rooms, setRooms] = useState<SocketRoom[]>([]);

  useEffect(() => {
    if (!isAuthenticated) {
      setStatus('disconnected');
      return;
    }

    if (status === 'connected' || status === 'connecting') {
      return;
    }

    const currentSocket = socket.current;

    if (
      currentSocket &&
      (currentSocket.readyState === WebSocket.OPEN || currentSocket.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    setStatus('connecting');
    const newSocket = new WebSocket(`${import.meta.env.VITE_API_URL}/ws?jwt=${tokenService.token}`);

    newSocket.onopen = () => {
      newSocket.send(JSON.stringify({ type: 'ping', data: null }));
      setStatus('connected');
    };

    newSocket.onclose = () => {
      setStatus('disconnected');
    };

    newSocket.onerror = () => {
      setStatus('disconnected');
    };

    socket.current = newSocket;

    return () => {
      if (
        socket.current &&
        (socket.current.readyState === WebSocket.OPEN || socket.current.readyState === WebSocket.CONNECTING)
      ) {
        return;
      }

      newSocket.close();
      socket.current = null;
    };
  }, [isAuthenticated]);

  useEffect(() => {
    if (status === 'connected' && socket.current) {
      socket.current.onmessage = (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data) as SocketEvent;

          if (data.type === 'ping') {
            send({ type: 'pong', data: null });
            return;
          }

          handlers.forEach((item) => item.handler(data));
        } catch (error) {
          return;
        }
      };
    }
  }, [handlers, status]);

  const registerHandler = useCallback((handler: HandlerItem) => {
    setHandlers((prev) => {
      if (prev.find((item) => item.id === handler.id)) {
        return prev.map((item) => (item.id === handler.id ? handler : item));
      }

      return [...prev, handler];
    });
  }, []);

  const unregisterHandlers = useCallback((ids: string[]) => {
    setHandlers((prev) => prev.filter((item) => !ids.includes(item.id)));
  }, []);

  const send = useCallback(
    (payload: SocketPayload<any>) => {
      if (status === 'connected' && socket.current) {
        socket.current.send(JSON.stringify(payload));
      }
    },
    [status],
  );

  const connectToRoom = useCallback(
    (roomId: string, type: SocketRoomType) => {
      if (rooms.find((room) => room.id === roomId && room.type === type)) {
        return;
      }

      send({ type: 'connect_user_to_room', data: { room_id: roomId, type } });

      setRooms((prev) => {
        return [...prev, { id: roomId, type }];
      });
    },
    [send, rooms],
  );

  const disconnectFromRoom = useCallback(
    (roomId: string) => {
      send({ type: 'disconnect_user_from_room', data: { room_id: roomId } });

      setRooms((prev) => prev.filter((room) => room.id !== roomId));
    },
    [send],
  );

  const s = socket.current;

  return (
    <SocketContext.Provider
      value={{ status, socket: s, send, registerHandler, connectToRoom, disconnectFromRoom, unregisterHandlers }}
    >
      {children}
    </SocketContext.Provider>
  );
};

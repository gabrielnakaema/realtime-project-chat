import { createContext, useCallback, useEffect, useEffectEvent, useRef, useState } from 'react';
import { CircleCheck, Loader2 } from 'lucide-react';
import type { SocketEvent } from '@/types/websocket';
import { useAuth } from '@/hooks/use-auth';
import { tokenService } from '@/services/api';

type WebSocketStatus = 'disconnected' | 'connected';

type SocketRoomType = 'chat' | 'project' | 'user' | '';

type SocketHandler = (event: SocketEvent) => void;

type ConnectionNotice = { type: 'hidden' } | { type: 'reconnecting'; seconds: number } | { type: 'reconnected' };

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
const RECONNECTED_NOTICE_DURATION_MS = 3_000;

export const SocketContext = createContext<SocketContextData>({} as SocketContextData);

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuth();
  const socket = useRef<WebSocket>(null);

  const [status, setStatus] = useState<WebSocketStatus>('disconnected');

  const [connectionAttempt, setConnectionAttempt] = useState(0);
  const subscriptions = useRef<Map<string, Subscription>>(new Map());
  const reconnectAttempt = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectCountdownTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const reconnectedNoticeTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnecting = useRef(false);
  const [connectionNotice, setConnectionNotice] = useState<ConnectionNotice>({ type: 'hidden' });

  const clearReconnectCountdown = useCallback(() => {
    if (reconnectCountdownTimer.current) {
      clearInterval(reconnectCountdownTimer.current);
      reconnectCountdownTimer.current = null;
    }
  }, []);

  const clearReconnectedNotice = useCallback(() => {
    if (reconnectedNoticeTimer.current) {
      clearTimeout(reconnectedNoticeTimer.current);
      reconnectedNoticeTimer.current = null;
    }
  }, []);

  const send = useCallback((payload: SocketPayload<any>) => {
    if (socket.current?.readyState === WebSocket.OPEN) {
      socket.current.send(JSON.stringify(payload));
    }
  }, []);

  const handleOpen = useEffectEvent(() => {
    const didReconnect = reconnecting.current;
    reconnecting.current = false;
    reconnectAttempt.current = 0;
    clearReconnectCountdown();
    clearReconnectedNotice();
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
    if (didReconnect) {
      setConnectionNotice({ type: 'reconnected' });
      reconnectedNoticeTimer.current = setTimeout(() => {
        reconnectedNoticeTimer.current = null;
        setConnectionNotice({ type: 'hidden' });
      }, RECONNECTED_NOTICE_DURATION_MS);
    } else {
      setConnectionNotice({ type: 'hidden' });
    }
  });

  const handleClose = useEffectEvent(() => {
    setStatus('disconnected');

    if (reconnectTimer.current) {
      return;
    }

    const delay = Math.min(RECONNECT_BASE_DELAY_MS * 2 ** reconnectAttempt.current, RECONNECT_MAX_DELAY_MS);
    reconnectAttempt.current += 1;
    reconnecting.current = true;
    clearReconnectCountdown();
    clearReconnectedNotice();
    let remainingSeconds = Math.ceil(delay / 1_000);
    setConnectionNotice({ type: 'reconnecting', seconds: remainingSeconds });
    if (remainingSeconds > 1) {
      reconnectCountdownTimer.current = setInterval(() => {
        remainingSeconds -= 1;
        setConnectionNotice({ type: 'reconnecting', seconds: remainingSeconds });
        if (remainingSeconds === 1) {
          clearReconnectCountdown();
        }
      }, 1_000);
    }
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
  }, [isAuthenticated, connectionAttempt, clearReconnectCountdown, clearReconnectedNotice]);

  useEffect(
    () => () => {
      reconnectAttempt.current = 0;
      reconnecting.current = false;
      clearReconnectCountdown();
      clearReconnectedNotice();
      setConnectionNotice({ type: 'hidden' });
    },
    [isAuthenticated, clearReconnectCountdown, clearReconnectedNotice],
  );

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

  const visibleConnectionNotice = isAuthenticated ? connectionNotice : { type: 'hidden' as const };

  return (
    <SocketContext.Provider value={{ status, subscribe }}>
      {children}
      {visibleConnectionNotice.type !== 'hidden' ? (
        <div
          aria-live="polite"
          className="pointer-events-none fixed bottom-4 left-4 z-50 flex items-center gap-2 rounded-lg border border-slate-200 bg-white/95 px-3 py-2 text-sm shadow-md backdrop-blur-sm dark:border-slate-700 dark:bg-slate-900/95"
          role="status"
        >
          {visibleConnectionNotice.type === 'reconnecting' ? (
            <>
              <Loader2 aria-hidden="true" className="size-4 animate-spin text-amber-500" />
              <span className="text-slate-700 dark:text-slate-200">
                Reconnecting in {visibleConnectionNotice.seconds}{' '}
                {visibleConnectionNotice.seconds === 1 ? 'second' : 'seconds'}…
              </span>
            </>
          ) : (
            <>
              <CircleCheck aria-hidden="true" className="animate-in zoom-in-50 size-4 text-emerald-500 duration-300" />
              <span className="text-slate-700 dark:text-slate-200">Connection restored</span>
            </>
          )}
        </div>
      ) : null}
    </SocketContext.Provider>
  );
};

import type { ChatMessage } from './chat';
import type { Task } from './task';

export type MessageEvent = {
  type: 'message';
  room_id: string;
  data: ChatMessage;
};

export type ErrorEvent = {
  type: 'error';
  room_id: string;
  data: {
    message: string;
  };
};

export type PingEvent = {
  type: 'ping';
  room_id: string;
  data: null;
};

export type TaskCreatedEvent = {
  type: 'task_created';
  room_id: string;
  data: Task;
};

export type TaskUpdatedEvent = {
  type: 'task_updated';
  room_id: string;
  data: Task;
};

export type UsersOnlineEvent = {
  type: 'users_online';
  room_id: string;
  data: string[];
};

export type UserDisconnectedEvent = {
  type: 'user_disconnected';
  room_id: string;
  data: {
    user_id: string;
    room_id: string;
  };
};

export type UserConnectedEvent = {
  type: 'user_connected';
  room_id: string;
  data: {
    user_id: string;
    room_id: string;
  };
};

export type SocketEvent =
  | MessageEvent
  | ErrorEvent
  | PingEvent
  | TaskCreatedEvent
  | TaskUpdatedEvent
  | UsersOnlineEvent
  | UserConnectedEvent
  | UserDisconnectedEvent;

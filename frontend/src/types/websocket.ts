import type { ChatMessage } from './chat';
import type { Task } from './task';

export type MessageEvent = {
  type: 'message';
  data: ChatMessage;
};

export type ErrorEvent = {
  type: 'error';
  data: {
    message: string;
  };
};

export type PingEvent = {
  type: 'ping';
  data: null;
};

export type TaskCreatedEvent = {
  type: 'task_created';
  data: Task;
};

export type TaskUpdatedEvent = {
  type: 'task_updated';
  data: Task;
};

export type SocketEvent = MessageEvent | ErrorEvent | PingEvent | TaskCreatedEvent | TaskUpdatedEvent;

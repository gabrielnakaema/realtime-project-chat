import type { ChatMessage } from './chat';

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

export type SocketEvent = MessageEvent | ErrorEvent | PingEvent;

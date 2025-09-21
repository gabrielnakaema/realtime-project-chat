import type { User } from './user';

export interface Chat {
  id: string;
  project_id: string;
  created_at: string;
  updated_at: string;
  members: ChatMember[];
  messages: ChatMessage[];
}

export interface ChatMember {
  id: string;
  user_id: string;
  chat_id: string;
  last_seen_at: string;
  joined_at: string;
  user: User | null;
}

type MessageType = 'text' | 'system';

export interface ChatMessage {
  id: string;
  message_type: MessageType;
  chat_id: string;
  user_id: string;
  content: string;
  created_at: string;
  updated_at: string;
  member: ChatMember | null;
}

export type SocketEvent =
  | {
      type: 'message';
      data: ChatMessage;
    }
  | {
      type: 'error';
      data: {
        message: string;
      };
    };

import type { User } from './user';

export interface Chat {
  id: string;
  project_id: string | null;
  chat_type: string;
  created_at: string;
  updated_at: string;
  unread_count: number;
  has_more_unread: boolean;
  members: ChatMember[];
  messages?: ChatMessage[];
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
  user_id: string | null;
  content: string;
  created_at: string;
  updated_at: string;
  reads_count: number;
  member: ChatMember | null;
}

export interface ChatMessageRead {
  message_id: string;
  user_id: string;
  read_at: string;
  user: User | null;
}

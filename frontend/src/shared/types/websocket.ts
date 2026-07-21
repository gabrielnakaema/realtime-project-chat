import type { ChatMessage, ChatMessageRead } from '@/features/chat/types/chat';
import type { Notification } from '@/features/notifications/types/notification';
import type { Member, Project } from '@/features/projects/types/project';
import type { Task, TaskComment } from '@/features/tasks/types/task';

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
  data: {
    task: Task;
    previous_project_column_id?: string;
  };
};

export type TaskCommentCreatedEvent = {
  type: 'task_comment_created';
  room_id: string;
  data: TaskComment;
};

export type ProjectUpdatedEvent = {
  type: 'project_updated';
  room_id: string;
  data: Project;
};

export type ProjectDeletedEvent = {
  type: 'project_deleted';
  room_id: string;
  data: Project;
};

export type ProjectMemberCreatedEvent = {
  type: 'project_member_created';
  room_id: string;
  data: Member;
};

export type ProjectMemberRemovedEvent = {
  type: 'project_member_removed';
  room_id: string;
  data: Member;
};

export type UsersOnlineEvent = {
  type: 'users_online';
  room_id: string;
  data: string[];
};

export type MessageReadEvent = {
  type: 'message_read';
  room_id: string;
  data: ChatMessageRead;
};

export type NotificationCreatedEvent = {
  type: 'notification_created';
  room_id: string;
  data: Notification;
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
  | TaskCommentCreatedEvent
  | ProjectUpdatedEvent
  | ProjectDeletedEvent
  | ProjectMemberCreatedEvent
  | ProjectMemberRemovedEvent
  | UsersOnlineEvent
  | MessageReadEvent
  | NotificationCreatedEvent
  | UserConnectedEvent
  | UserDisconnectedEvent;

export type ProjectMembershipEvent = ProjectMemberCreatedEvent | ProjectMemberRemovedEvent;

export const isProjectMembershipEvent = (event: SocketEvent): event is ProjectMembershipEvent =>
  event.type === 'project_member_created' || event.type === 'project_member_removed';

export const isProjectDeletedEvent = (event: SocketEvent): event is ProjectDeletedEvent =>
  event.type === 'project_deleted';

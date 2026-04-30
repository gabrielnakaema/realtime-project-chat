import type { ListTasksRequest } from '@/types/task';
import type { SearchTasksForUserRequest } from './tasks';

export const userQueryKeys = {
  me: ['users', 'me'],
} as const;

export const projectQueryKeys = {
  all: ['projects'],
  list: ['projects', 'list'],
  search: (query?: string) => ['projects', 'list', { query }],
  details: (id: string) => ['projects', 'details', id],
  members: (id: string) => ['projects', 'members', id],
  activities: (id: string) => ['projects', 'activities', id],
} as const;

export const taskQueryKeys = {
  all: ['tasks'],
  listByProjectId: (projectId: string) => ['tasks', 'list', { projectId }],
  details: (id: string) => ['tasks', 'details', id],
  comments: (taskId: string) => ['tasks', 'comments', taskId],
  list: (request: ListTasksRequest) => ['tasks', 'list', request],
  listGroupedByProjectId: (request: ListTasksRequest) => ['tasks', 'list', 'grouped', request],
  countByColumn: (projectId: string, projectColumnIds: string[]) => ['tasks', 'count', { projectId, projectColumnIds }],
  listUserDueTasks: ['tasks', 'list', 'user'],
  search: (request: SearchTasksForUserRequest) => ['tasks', 'search', request],

  _allGrouped: () => ['tasks', 'list', 'grouped'] as const,
  _allCounts: () => ['tasks', 'count'] as const,
} as const;

export const chatMessageQueryKeys = {
  reads: (chatId: string, messageId: string) => ['chats', chatId, 'message-reads', messageId],
} as const;

export const projectChatQueryKeys = {
  all: ['chats'],
  detailsByProjectId: (projectId: string) => ['chats', 'details', { projectId }],
  infiniteMessages: (projectId: string) => ['chats', 'messages', 'infinite', { projectId }],
  listMessagesByProjectId: ({ projectId, before }: { projectId: string; before?: string }) => [
    'chats',
    'messages',
    { projectId, before },
  ],
} as const;

export const generalChatQueryKeys = {
  list: ['general-chats'],
  details: (chatId: string) => ['general-chats', chatId],
  infiniteMessages: (chatId: string) => ['general-chats', chatId, 'messages'],
} as const;

export const notificationQueryKeys = {
  all: ['notifications'],
  list: ['notifications', 'list'],
  unreadCount: ['notifications', 'unread-count'],
} as const;

export const userListQueryKeys = {
  all: ['users', 'list'],
} as const;

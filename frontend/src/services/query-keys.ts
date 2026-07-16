import type {
  SearchProjectTasksForDependenciesRequest,
  SearchTasksForUserRequest,
  SuggestTaskCodesRequest,
} from './tasks';

export interface TaskCommentsQueryKeyState {
  taskId: string;
  commentId: string | null;
  commentCreatedAt: string | null;
  limit: number | null;
}

export type TaskCommentsQueryKey = readonly ['tasks', 'comments', TaskCommentsQueryKeyState];

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
  comments: (
    taskId: string,
    options?: {
      commentId?: string | null;
      commentCreatedAt?: string | null;
      limit?: number | null;
    },
  ): TaskCommentsQueryKey => [
    'tasks',
    'comments',
    {
      taskId,
      commentId: options?.commentId ?? null,
      commentCreatedAt: options?.commentCreatedAt ?? null,
      limit: options?.limit ?? null,
    },
  ],

  board: (projectId: string) => ['tasks', 'board', projectId] as const,
  boardColumn: (projectId: string, columnId: string) => ['tasks', 'board', projectId, columnId] as const,
  counts: (projectId: string) => ['tasks', 'count', projectId] as const,
  countsFor: (projectId: string, columnIds: string[]) => ['tasks', 'count', projectId, { columnIds }] as const,
  archived: (projectId: string) => ['tasks', 'archived', projectId] as const,
  listUserDueTasks: ['tasks', 'list', 'user'],
  search: (request: SearchTasksForUserRequest) => ['tasks', 'search', request],
  searchProjectDependencies: (request: SearchProjectTasksForDependenciesRequest) => [
    'tasks',
    'project-search',
    request,
  ],
  codeSuggestions: (request: SuggestTaskCodesRequest) => ['tasks', 'code-suggestions', request],
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

export const mcpAPIKeyQueryKeys = {
  all: ['mcp-api-keys'],
  list: ['mcp-api-keys', 'list'],
  scopes: ['mcp-api-keys', 'scopes'],
} as const;

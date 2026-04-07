import type { ListTasksRequest, TaskStatus } from '@/types/task';
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
  countByStatus: (projectId: string, statuses: TaskStatus[]) => ['tasks', 'count', { projectId, statuses }],
  listUserDueTasks: ['tasks', 'list', 'user'],
  search: (request: SearchTasksForUserRequest) => ['tasks', 'search', request],

  _allGrouped: () => ['tasks', 'list', 'grouped'] as const,
  _allCounts: () => ['tasks', 'count'] as const,
} as const;

export const chatQueryKeys = {
  all: ['chats'],
  detailsByProjectId: (projectId: string) => ['chats', 'details', { projectId }],
  listInfiniteMessagesByProjectId: ({ projectId }: { projectId: string }) => [
    'chats',
    'messages',
    'infinite',
    { projectId },
  ],
  listMessagesByProjectId: ({ projectId, before }: { projectId: string; before?: string }) => [
    'chats',
    'messages',
    { projectId, before },
  ],
} as const;

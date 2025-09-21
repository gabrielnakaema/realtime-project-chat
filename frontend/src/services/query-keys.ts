export const userQueryKeys = {
  me: ['users', 'me'],
} as const;

export const projectQueryKeys = {
  all: ['projects'],
  list: ['projects', 'list'],
  details: (id: string) => ['projects', 'details', id],
} as const;

export const taskQueryKeys = {
  all: ['tasks'],
  listByProjectId: (projectId: string) => ['tasks', 'list', { projectId }],
  details: (id: string) => ['tasks', 'details', id],
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

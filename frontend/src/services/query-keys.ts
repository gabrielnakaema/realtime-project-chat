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

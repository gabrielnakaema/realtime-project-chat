import type { TaskStatus } from '@/types/task';

export const TASK_STATUSES: readonly TaskStatus[] = ['pending', 'doing', 'done'] as const;

export const DEFAULT_TASK_LIMIT = 15;

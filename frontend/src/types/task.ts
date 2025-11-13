import type { User } from './user';

export type TaskStatus = 'pending' | 'doing' | 'done' | 'archived';

export interface Task {
  id: string;
  project_id: string;
  title: string;
  description: string;
  status: TaskStatus;
  created_at: string;
  updated_at: string;
  author_id: string;
  author: User;
  changes: TaskChange[];
}

export interface TaskChange {
  id: string;
  task_id: string;
  author_id: string;
  change_description: string;
  created_at: string;
  author: User | null;
}

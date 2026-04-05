import type { Project } from './project';
import type { User } from './user';

export type TaskStatus = 'pending' | 'doing' | 'done' | 'archived';

export type TaskPriority = 'low' | 'medium' | 'high';

export interface Task {
  id: string;
  project_id: string;
  title: string;
  description: string;
  status: TaskStatus;
  created_at: string;
  updated_at: string;
  priority: TaskPriority;
  order: string;
  responsible_id: string | null;
  due_date: string | null;
  done_at: string | null;
  tags: string[] | null;
  author_id: string;
  author: User;
  responsible: User | null;
  updates: TaskUpdate[];
  project: Project | null;
}

export interface TaskUpdate {
  id: string;
  task_id: string;
  user_id: string;
  update_type: string;
  created_at: string;
  user: User;
  changes: TaskChange[] | null;
}

export interface TaskChange {
  id: string;
  update_id: string;
  subject_id: string | null;
  old_value_id: string | null;
  new_value_id: string | null;
  old_display_value?: string | null;
  new_display_value?: string | null;
  field: string;
  old_value: string;
  new_value: string;
  created_at: string;
  subject: User | null;
}

export interface ListTasksRequest {
  projectId: string;
  statuses: string[];
  taskOrder: string;
  updatedAt: null | string;
  limit: number;
}

export interface ListUserDueTasksRequest {
  cursorDueDate: null | string;
  cursorUpdatedAt: null | string;
  limit: number;
}

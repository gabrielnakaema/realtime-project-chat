import type { ActionOrigin } from './action-origin';
import type { Project, ProjectColumn } from './project';
import type { User } from './user';

export type TaskStatus = string;

export type TaskPriority = 'low' | 'medium' | 'high';

export interface TaskDependencyRef {
  id: string;
  title: string;
  code: string;
}

export interface Task {
  id: string;
  project_id: string;
  title: string;
  description: string;
  code: string | null;
  status: TaskStatus;
  project_column_id: string;
  project_column: ProjectColumn | null;
  created_at: string;
  updated_at: string;
  priority: TaskPriority;
  order: string;
  responsible_id: string | null;
  due_date: string | null;
  done_at: string | null;
  archived_at: string | null;
  tags: string[] | null;
  depends_on_task_ids?: string[];
  depends_on_tasks?: TaskDependencyRef[];
  author_id: string;
  author: User;
  responsible: User | null;
  updates: TaskUpdate[];
  project: Project | null;
}

export interface TaskComment {
  id: string;
  task: Task | null;
  user: User;
  content: string;
  parent_comment_id?: string | null;
  action_origin?: ActionOrigin | null;
  created_at: string;
  updated_at: string;
  replies: TaskComment[];
}

export interface TaskUpdate {
  id: string;
  task_id: string;
  user_id: string;
  update_type: string;
  action_origin?: ActionOrigin | null;
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
  projectColumnIds: string[];
  archived: boolean;
  taskOrder: string;
  updatedAt: null | string;
  limit: number;
}

export interface ListUserDueTasksRequest {
  cursorDueDate: null | string;
  cursorUpdatedAt: null | string;
  limit: number;
}

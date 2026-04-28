import type { Project } from './project';
import type { Task, TaskComment } from './task';
import type { User } from './user';

export type NotificationType = 'project.member.created' | 'task.assigned' | 'task.comment.created';

export interface Notification {
  id: string;
  user_id: string;
  actor_id: string;
  project_id: string;
  task_id: string | null;
  task_comment_id: string | null;
  type: NotificationType;
  read_at: string | null;
  created_at: string;
  updated_at: string;
  actor: User | null;
  project: Project | null;
  task: Task | null;
  task_comment: TaskComment | null;
}

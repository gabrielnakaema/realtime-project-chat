import type { User } from './user';

export interface Task {
  id: string;
  project_id: string;
  title: string;
  description: string;
  status: string;
  created_at: string;
  updated_at: string;
  author_id: string;
  author: User;
}

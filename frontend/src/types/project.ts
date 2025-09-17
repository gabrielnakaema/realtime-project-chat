import type { User } from './user';

export interface Project {
  id: string;
  user_id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  members: Member[];
}

export interface Member {
  id: string;
  user_id: string;
  user: User;
  project_id: string;
  role: string;
}

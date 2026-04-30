import type { User } from './user';

export interface Project {
  id: string;
  user_id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  members: Member[];
  columns: ProjectColumn[];
}

export interface ProjectColumn {
  id?: string;
  project_id?: string;
  name: string;
  color: string;
  position: number;
  is_done_column: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface Member {
  id: string;
  user_id: string;
  user: User;
  project_id: string;
  role: string;
}

export enum ProjectMemberRole {
  Creator = 'creator',
  Member = 'member',
}

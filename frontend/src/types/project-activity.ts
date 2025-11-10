import type { Member, Project } from './project';
import type { Task } from './task';
import type { User } from './user';

export type ProjectActivity =
  | TaskProjectActivity
  | ProjectMemberProjectActivity
  | ProjectProjectActivity
  | UnknownProjectActivity;

type ProjectActivityEntityType = 'task' | 'project' | 'project.member' | '';

type ProjectActivityType =
  | 'task.created'
  | 'task.updated'
  | 'task.deleted'
  | 'project.created'
  | 'project.updated'
  | 'project.member.created'
  | 'project.member.deleted';

interface BaseProjectActivity {
  id: string;
  project: Project | null;
  actor: User | null;
  activity_type: ProjectActivityType;
  activity_data: any;
  entity_id: string;
  created_at: string;
  updated_at: string;
  entity_type: ProjectActivityEntityType;
}

export interface TaskProjectActivity extends BaseProjectActivity {
  entity_type: 'task';
  task: Task | null;
}

export interface ProjectMemberProjectActivity extends BaseProjectActivity {
  entity_type: 'project.member';
  project_member: Member | null;
}

export interface ProjectProjectActivity extends BaseProjectActivity {
  entity_type: 'project';
}

export interface UnknownProjectActivity extends BaseProjectActivity {
  entity_type: '';
}

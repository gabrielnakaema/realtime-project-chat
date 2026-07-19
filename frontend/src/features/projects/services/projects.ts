import type { Member, Project, ProjectColumn } from '@/features/projects/types/project';
import type { IProjectForm } from '@/features/projects/schemas/project.schema';
import { api } from '@/shared/services/api';

export const listProjects = async (searchQuery?: string) => {
  const searchParams = new URLSearchParams();
  if (searchQuery) {
    searchParams.set('query', searchQuery.toString());
  }

  const response = await api.get('projects', {
    searchParams,
  });

  const json = await response.json<Project[]>();
  return json;
};

export const getProject = async (projectId: string) => {
  const response = await api.get(`projects/${projectId}`);

  const json = await response.json<Project>();
  return json;
};

export const createProject = async (form: IProjectForm) => {
  const response = await api.post('projects', {
    json: form,
  });

  const json = await response.json<Project>();

  return json;
};

interface UpdateProjectRequest {
  id: string;
  name: string;
  description: string;
  repository_url: string;
  repository_owner: string;
  repository_name: string;
  default_branch: string;
  branch_name_prefix: string;
  columns: IProjectForm['columns'];
  deleted_columns: IProjectForm['deleted_columns'];
}

export const updateProject = async (request: UpdateProjectRequest) => {
  const payload = {
    name: request.name,
    description: request.description,
    repository_url: request.repository_url,
    repository_owner: request.repository_owner,
    repository_name: request.repository_name,
    default_branch: request.default_branch,
    branch_name_prefix: request.branch_name_prefix,
    columns: request.columns,
    deleted_columns: request.deleted_columns,
  };

  const response = await api.put(`projects/${request.id}`, {
    json: payload,
  });

  const json = await response.json<Project>();

  return json;
};

interface UpdateProjectColumnRequest {
  projectId: string;
  columnId: string;
  name: string;
  description: string;
  color: string;
  is_done_column: boolean;
}

export const updateProjectColumn = async (request: UpdateProjectColumnRequest) => {
  const payload = {
    name: request.name,
    description: request.description,
    color: request.color,
    is_done_column: request.is_done_column,
  };

  const response = await api.patch(`projects/${request.projectId}/columns/${request.columnId}`, {
    json: payload,
  });

  const json = await response.json<ProjectColumn>();

  return json;
};

interface CreateProjectMemberRequest {
  projectId: string;
  email: string;
}

export const createProjectMember = async (request: CreateProjectMemberRequest) => {
  const payload = {
    email: request.email,
  };

  const response = await api.post(`projects/${request.projectId}/members`, {
    json: payload,
  });

  const json = await response.json();
  return json;
};

export const listMembersByProjectId = async (projectId: string) => {
  const response = await api.get(`projects/${projectId}/members`);

  const json = await response.json<Member[]>();
  return json;
};

interface RemoveProjectMemberRequest {
  projectId: string;
  memberUserId: string;
}

export const removeProjectMember = async (request: RemoveProjectMemberRequest) => {
  const response = await api.delete(`projects/${request.projectId}/members/${request.memberUserId}`);

  return response.ok;
};

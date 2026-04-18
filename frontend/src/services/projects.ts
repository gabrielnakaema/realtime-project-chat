import { api } from './api';
import type { Member, Project } from '@/types/project';
import type { IProjectForm } from '@/schemas/project-schema';

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
}

export const updateProject = async (request: UpdateProjectRequest) => {
  const payload = {
    name: request.name,
    description: request.description,
  };

  const response = await api.put(`projects/${request.id}`, {
    json: payload,
  });

  const json = await response.json<Project>();

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

import { api } from './api';
import type { Project } from '@/types/project';
import type { IProjectForm } from '@/schemas/project-schema';

export const listProjects = async () => {
  const response = await api.get('projects');

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

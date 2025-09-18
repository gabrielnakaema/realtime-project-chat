import type { Project } from '@/types/project';
import { api } from './api';
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

interface CreateProjectMemberRequest {
  projectId: string;
  email: string;
}

export const createProjectMember = async (form: CreateProjectMemberRequest) => {
  const payload = {
    email: form.email,
  };

  const response = await api.post(`projects/${form.projectId}/members`, {
    json: payload,
  });

  const json = await response.json();
  return json;
};

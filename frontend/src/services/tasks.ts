import type { Task } from '@/types/task';
import { api } from './api';
import type { ITaskForm } from '@/schemas/task-schema';
import type { Paginated } from '@/types/paginated';

export const listTasksByProjectId = async (projectId: string) => {
  const response = await api.get(`tasks?project_id=${projectId}`);

  const json = await response.json<Paginated<Task>>();
  return json;
};

interface CreateTaskRequest {
  projectId: string;
  form: ITaskForm;
}

export const createTask = async (request: CreateTaskRequest) => {
  const payload = {
    project_id: request.projectId,
    title: request.form.title,
    description: request.form.description,
  };

  const response = await api.post('tasks', {
    json: payload,
  });

  const json = await response.json<Task>();
  return json;
};

interface UpdateTaskRequest {
  id: string;
  title: string;
  description: string;
  status: string;
}

export const updateTask = async (request: UpdateTaskRequest) => {
  const payload = {
    title: request.title,
    description: request.description,
    status: request.status,
  };

  const response = await api.put(`tasks/${request.id}`, {
    json: payload,
  });

  const json = await response.json<Task>();
  return json;
};

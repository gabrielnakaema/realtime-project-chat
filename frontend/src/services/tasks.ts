import { parse } from 'date-fns';
import { api } from './api';
import type { ListTasksRequest, Task } from '@/types/task';
import type { ITaskForm } from '@/schemas/task-schema';
import type { Paginated } from '@/types/paginated';

export const listTasksByProjectId = async (request: ListTasksRequest) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', request.projectId);
  if (request.statuses.length > 0) {
    searchParams.set('statuses', request.statuses.join(','));
  }
  if (request.taskOrder) {
    searchParams.set('task_order', request.taskOrder.toString());
  }
  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }

  const response = await api.get('tasks', {
    searchParams,
  });

  const json = await response.json<Paginated<Task>>();
  return json;
};

const formatDateForApi = (date: string | null | undefined) => {
  if (!date) return null;

  try {
    if (date.includes('T')) {
      return date;
    }

    return parse(date, 'yyyy-MM-dd', new Date()).toISOString();
  } catch (error) {
    return null;
  }
};

interface CreateTaskRequest {
  projectId: string;
  form: ITaskForm;
}

export const createTask = async (request: CreateTaskRequest) => {
  const formattedDueDate = formatDateForApi(request.form.due_date);

  const formTags = request.form.tags ? request.form.tags.split(',').map((tag) => tag.trim()) : null;
  const uniqueTags = formTags?.length ? Array.from(new Set(formTags)) : null;

  const payload = {
    project_id: request.projectId,
    title: request.form.title,
    description: request.form.description,
    priority: request.form.priority,
    responsible_id: request.form.responsible_id,
    due_date: formattedDueDate,
    tags: uniqueTags,
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

  order: number;
  priority: string;
  due_date: string | null;
  responsible_id: string | null;
  done_at: string | null;
  tags: string[];
}

export const updateTask = async (request: UpdateTaskRequest) => {
  const formattedDueDate = formatDateForApi(request.due_date);

  const formattedDoneAt = formatDateForApi(request.done_at);

  const payload = {
    title: request.title,
    description: request.description,
    status: request.status,
    order: request.order,
    priority: request.priority,
    due_date: formattedDueDate,
    responsible_id: request.responsible_id || null,
    done_at: formattedDoneAt,
    tags: request.tags,
  };

  const response = await api.put(`tasks/${request.id}`, {
    json: payload,
  });

  const json = await response.json<Task>();
  return json;
};

export const getTask = async (taskId: string) => {
  const response = await api.get(`tasks/${taskId}`);

  const json = await response.json<Task>();
  return json;
};

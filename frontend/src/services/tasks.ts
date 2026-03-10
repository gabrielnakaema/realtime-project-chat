import { parse } from 'date-fns';
import { api } from './api';
import type { ListTasksRequest, ListUserDueTasksRequest, Task, TaskStatus } from '@/types/task';
import type { ITaskForm } from '@/schemas/task-schema';
import type { CursorPaginated, Paginated } from '@/types/paginated';

export const countTasksByStatus = async (projectId: string, statuses: TaskStatus[]) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', projectId);
  if (statuses.length > 0) {
    searchParams.set('statuses', statuses.join(','));
  }

  const response = await api.get('tasks/count-by-status', {
    searchParams,
  });

  const json = await response.json<Record<TaskStatus, number>>();
  return json;
};

export const listGroupedTasksByProjectId = async (request: ListTasksRequest) => {
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

  if (request.updatedAt) {
    searchParams.set('updated_at', request.updatedAt);
  }

  const response = await api.get('tasks/group-by-status', {
    searchParams,
  });

  const json = await response.json<Record<TaskStatus, CursorPaginated<Task>>>();
  return json;
};

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

export const archiveTask = async (taskId: string) => {
  const response = await api.delete(`tasks/${taskId}`);

  const json = await response.json<Task>();
  return json;
};

export const getTask = async (taskId: string) => {
  const response = await api.get(`tasks/${taskId}`);

  const json = await response.json<Task>();
  return json;
};

interface MoveTaskRequest {
  taskId: string;
  projectId: string;
  status: TaskStatus;
  afterTaskId: string | null;
}

export const moveTask = async (request: MoveTaskRequest) => {
  const payload = {
    after_task_id: request.afterTaskId,
    project_id: request.projectId,
    status: request.status,
  };

  const response = await api.patch(`tasks/${request.taskId}/move`, {
    json: payload,
  });

  const json = await response.json<Task>();
  return json;
};

export const listUserDueTasks = async (request: ListUserDueTasksRequest) => {
  const searchParams = new URLSearchParams();
  if (request.cursorDueDate) {
    searchParams.set('due_date', request.cursorDueDate);
  }
  if (request.cursorUpdatedAt) {
    searchParams.set('updated_at', request.cursorUpdatedAt);
  }
  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }

  const response = await api.get('tasks/user', {
    searchParams,
  });

  const json = await response.json<CursorPaginated<Task>>();
  return json;
};

export interface SearchTasksForUserRequest {
  cursorDueDate: null | string;
  cursorUpdatedAt: null | string;
  limit: number;
  searchQuery: string;
}

export const searchTasksForUser = async (request: SearchTasksForUserRequest) => {
  const searchParams = new URLSearchParams();
  if (request.cursorDueDate) {
    searchParams.set('due_date', request.cursorDueDate);
  }
  if (request.cursorUpdatedAt) {
    searchParams.set('updated_at', request.cursorUpdatedAt);
  }
  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }
  if (request.searchQuery) {
    searchParams.set('query', request.searchQuery);
  }

  const response = await api.get('tasks/search', {
    searchParams,
  });

  const json = await response.json<CursorPaginated<Task>>();
  return json;
};

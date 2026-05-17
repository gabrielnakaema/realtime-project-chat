import { parse } from 'date-fns';
import { api } from './api';
import type { ListTasksRequest, ListUserDueTasksRequest, Task, TaskComment } from '@/types/task';
import type { ITaskForm } from '@/schemas/task-schema';
import type { CursorPaginated, Paginated } from '@/types/paginated';

export const countTasksByColumn = async (projectId: string, projectColumnIds: string[]) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', projectId);
  if (projectColumnIds.length > 0) {
    searchParams.set('project_column_ids', projectColumnIds.join(','));
  }

  const response = await api.get('tasks/count-by-column', {
    searchParams,
  });

  const json = await response.json<Record<string, number>>();
  return json;
};

export const listGroupedTasksByProjectId = async (request: ListTasksRequest) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', request.projectId);
  if (request.projectColumnIds.length > 0) {
    searchParams.set('project_column_ids', request.projectColumnIds.join(','));
  }
  searchParams.set('archived', String(request.archived));
  if (request.taskOrder) {
    searchParams.set('task_order', request.taskOrder);
  }

  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }

  if (request.updatedAt) {
    searchParams.set('updated_at', request.updatedAt);
  }

  const response = await api.get('tasks/group-by-column', {
    searchParams,
  });

  const json = await response.json<Record<string, CursorPaginated<Task>>>();
  return json;
};

export const listTasksByProjectId = async (request: ListTasksRequest) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', request.projectId);
  if (request.projectColumnIds.length > 0) {
    searchParams.set('project_column_ids', request.projectColumnIds.join(','));
  }
  searchParams.set('archived', String(request.archived));
  if (request.taskOrder) {
    searchParams.set('task_order', request.taskOrder);
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
    project_column_id: request.form.project_column_id,
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
  project_column_id: string;

  priority: string;
  due_date: string | null;
  responsible_id: string | null;
  tags: string[];
}

export const updateTask = async (request: UpdateTaskRequest) => {
  const formattedDueDate = formatDateForApi(request.due_date);

  const payload = {
    title: request.title,
    description: request.description,
    project_column_id: request.project_column_id,
    priority: request.priority,
    due_date: formattedDueDate,
    responsible_id: request.responsible_id || null,
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

export const restoreTask = async (taskId: string, projectColumnId: string) => {
  const response = await api.post(`tasks/${taskId}/restore`, {
    json: {
      project_column_id: projectColumnId,
    },
  });

  const json = await response.json<Task>();
  return json;
};

export const getTask = async (taskId: string) => {
  const response = await api.get(`tasks/${taskId}`);

  const json = await response.json<Task>();
  return json;
};

interface ListTaskCommentsRequest {
  taskId: string;
  limit?: number;
  before?: string;
  commentId?: string;
}

export const listTaskComments = async (request: ListTaskCommentsRequest) => {
  const searchParams = new URLSearchParams();
  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }
  if (request.before) {
    searchParams.set('before', request.before);
  }
  if (request.commentId) {
    searchParams.set('comment_id', request.commentId);
  }

  const response = await api.get(`tasks/${request.taskId}/comments`, {
    searchParams,
  });

  const json = await response.json<CursorPaginated<TaskComment>>();
  return json;
};

interface CreateTaskCommentRequest {
  taskId: string;
  content: string;
  parentCommentId?: string | null;
}

export const createTaskComment = async (request: CreateTaskCommentRequest) => {
  const response = await api.post(`tasks/${request.taskId}/comments`, {
    json: {
      content: request.content,
      parent_comment_id: request.parentCommentId ?? null,
    },
  });

  const json = await response.json<TaskComment>();
  return json;
};

interface MoveTaskRequest {
  taskId: string;
  projectId: string;
  projectColumnId: string;
  projectColumnIds: string[];
  afterTaskId: string | null;
}

export const moveTask = async (request: MoveTaskRequest) => {
  const payload = {
    after_task_id: request.afterTaskId,
    project_id: request.projectId,
    project_column_id: request.projectColumnId,
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

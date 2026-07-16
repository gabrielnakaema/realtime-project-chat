import { parse } from 'date-fns';
import { api } from './api';
import type { ListUserDueTasksRequest, Task, TaskCodeSuggestion, TaskComment, TaskDependencyRef } from '@/types/task';
import type { ITaskForm } from '@/schemas/task-schema';
import type { CursorPaginated } from '@/types/paginated';

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

export interface ListColumnTasksRequest {
  projectId: string;
  columnId?: string;
  archived: boolean;
  limit: number;
  taskOrder?: string;
  updatedAt?: string | null;
}

export const listColumnTasks = async (request: ListColumnTasksRequest) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', request.projectId);
  if (request.columnId) {
    searchParams.set('project_column_ids', request.columnId);
  }
  searchParams.set('archived', String(request.archived));
  searchParams.set('limit', request.limit.toString());
  if (request.taskOrder) {
    searchParams.set('task_order', request.taskOrder);
  }
  if (request.updatedAt) {
    searchParams.set('updated_at', request.updatedAt);
  }

  const response = await api.get('tasks', { searchParams });

  const json = await response.json<CursorPaginated<Task>>();
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
  const normalizedCode = request.form.code?.trim() || null;

  const formTags = request.form.tags ? request.form.tags.split(',').map((tag) => tag.trim()) : null;
  const uniqueTags = formTags?.length ? Array.from(new Set(formTags)) : null;

  const payload = {
    project_id: request.projectId,
    project_column_id: request.form.project_column_id,
    title: request.form.title,
    description: request.form.description,
    code: normalizedCode,
    priority: request.form.priority,
    responsible_id: request.form.responsible_id,
    due_date: formattedDueDate,
    tags: uniqueTags,
    depends_on_task_ids: request.form.depends_on_task_ids.length > 0 ? request.form.depends_on_task_ids : null,
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
  code: string | null;
  project_column_id: string;

  priority: string;
  due_date: string | null;
  responsible_id: string | null;
  tags: string[];
  depends_on_task_ids: string[];
}

export const updateTask = async (request: UpdateTaskRequest) => {
  const formattedDueDate = formatDateForApi(request.due_date);
  const normalizedCode = request.code?.trim() || null;

  const payload = {
    title: request.title,
    description: request.description,
    code: normalizedCode,
    project_column_id: request.project_column_id,
    priority: request.priority,
    due_date: formattedDueDate,
    responsible_id: request.responsible_id || null,
    tags: request.tags,
    depends_on_task_ids: request.depends_on_task_ids,
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
  beforeCommentId?: string;
  after?: string;
  afterCommentId?: string;
}

export const listTaskComments = async (request: ListTaskCommentsRequest) => {
  const searchParams = new URLSearchParams();
  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }
  if (request.before) {
    searchParams.set('before', request.before);
  }
  if (request.beforeCommentId) {
    searchParams.set('comment_id', request.beforeCommentId);
  }
  if (request.after) {
    searchParams.set('after', request.after);
  }
  if (request.afterCommentId) {
    searchParams.set('after_comment_id', request.afterCommentId);
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

export interface SearchProjectTasksForDependenciesRequest {
  projectId: string;
  query: string;
  excludeTaskId?: string;
  limit?: number;
}

export interface SuggestTaskCodesRequest {
  projectId: string;
  prefix: string;
  limit?: number;
}

export const suggestTaskCodes = async (request: SuggestTaskCodesRequest) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', request.projectId);
  searchParams.set('prefix', request.prefix);

  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }

  const response = await api.get('tasks/code-suggestions', {
    searchParams,
  });

  const json = await response.json<{ data: TaskCodeSuggestion[] }>();
  return json.data;
};

export const searchProjectTasksForDependencies = async (request: SearchProjectTasksForDependenciesRequest) => {
  const searchParams = new URLSearchParams();
  searchParams.set('project_id', request.projectId);
  searchParams.set('query', request.query);

  if (request.excludeTaskId) {
    searchParams.set('exclude_task_id', request.excludeTaskId);
  }

  if (request.limit) {
    searchParams.set('limit', request.limit.toString());
  }

  const response = await api.get('tasks/project-search', {
    searchParams,
  });

  const json = await response.json<{ data: TaskDependencyRef[] }>();
  return json.data;
};

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

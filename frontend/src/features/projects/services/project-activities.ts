import type { CursorPaginated } from '@/shared/types/paginated';
import type { ProjectActivity } from '@/features/projects/types/project-activity';
import { api } from '@/shared/services/api';

interface ListUserProjectActivitiesPayload {
  before?: string;
  id?: string;
  limit?: number;
}

export const listUserProjectActivities = async (payload: ListUserProjectActivitiesPayload) => {
  const searchParams = new URLSearchParams();
  if (payload.before) {
    searchParams.set('before', payload.before);
  }

  if (payload.id) {
    searchParams.set('id', payload.id);
  }

  searchParams.set('limit', payload.limit?.toString() ?? '30');

  const response = await api.get('projects/activities', {
    searchParams,
  });

  const json = await response.json<CursorPaginated<ProjectActivity>>();

  return json;
};

interface ListProjectActivitiesPayload {
  projectId: string;
  before?: string;
  id?: string;
  limit?: number;
}

export const listProjectActivities = async (payload: ListProjectActivitiesPayload) => {
  const searchParams = new URLSearchParams();
  if (payload.before) {
    searchParams.set('before', payload.before);
  }

  if (payload.id) {
    searchParams.set('id', payload.id);
  }

  searchParams.set('limit', payload.limit?.toString() ?? '10');

  const response = await api.get(`projects/${payload.projectId}/activities`, {
    searchParams,
  });

  const json = await response.json<CursorPaginated<ProjectActivity>>();

  return json;
};

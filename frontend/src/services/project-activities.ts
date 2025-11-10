import { api } from './api';
import type { CursorPaginated } from '@/types/paginated';
import type { ProjectActivity } from '@/types/project-activity';

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

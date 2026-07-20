import type { InfiniteData, QueryClient } from '@tanstack/react-query';
import type { CursorPaginated } from '@/shared/types/paginated';
import type { Task } from '@/features/tasks/types/task';
import { taskQueryKeys } from '@/shared/services/query-keys';

export interface BoardColumnPage extends CursorPaginated<Task> {
  cursor: { taskOrder: string; updatedAt: string } | null;
}

export type ColumnCache = InfiniteData<BoardColumnPage>;

export function compareTasks(a: Task, b: Task): number {
  if (a.order !== b.order) return a.order < b.order ? -1 : 1;
  if (a.updated_at !== b.updated_at) return a.updated_at > b.updated_at ? -1 : 1;
  return a.id < b.id ? -1 : a.id > b.id ? 1 : 0;
}

export function sortedColumnTasks(cache: Pick<ColumnCache, 'pages'> | undefined): Task[] {
  if (!cache) return [];
  return cache.pages
    .flatMap((page) => page.data)
    .slice()
    .sort(compareTasks);
}

export function findTaskOnBoard(
  queryClient: QueryClient,
  projectId: string,
  taskId: string,
): { task: Task; columnId: string } | undefined {
  const entries = queryClient.getQueriesData<ColumnCache>({ queryKey: taskQueryKeys.board(projectId) });
  for (const [key, data] of entries) {
    if (!data) continue;

    const columnId = key[3];
    if (typeof columnId !== 'string') continue;

    for (const page of data.pages) {
      const found = page.data.find((t) => t.id === taskId);
      if (found) return { task: found, columnId };
    }
  }
  return undefined;
}

export function removeTask(queryClient: QueryClient, projectId: string, columnId: string, taskId: string) {
  queryClient.setQueryData<ColumnCache>(taskQueryKeys.boardColumn(projectId, columnId), (old) => {
    if (!old) return old;
    return {
      ...old,
      pages: old.pages.map((page) => ({ ...page, data: page.data.filter((t) => t.id !== taskId) })),
    };
  });
}

export function upsertTask(
  queryClient: QueryClient,
  projectId: string,
  columnId: string,
  task: Task,
  options: { allowBeyondWindow?: boolean } = {},
) {
  queryClient.setQueryData<ColumnCache>(taskQueryKeys.boardColumn(projectId, columnId), (old) => {
    if (!old || old.pages.length === 0) return old;

    const pages = old.pages.map((page) => ({ ...page, data: page.data.filter((t) => t.id !== task.id) }));

    const hasNext = pages[pages.length - 1].has_next;
    const lastOrder = pages
      .flatMap((page) => page.data)
      .reduce<string | undefined>((max, t) => (max === undefined || t.order > max ? t.order : max), undefined);
    const beyondWindow = hasNext && lastOrder !== undefined && task.order > lastOrder;

    if (beyondWindow && !options.allowBeyondWindow) {
      return { ...old, pages };
    }

    const lastPage = pages[pages.length - 1];
    pages[pages.length - 1] = { ...lastPage, data: [...lastPage.data, task] };

    return { ...old, pages };
  });
}

const seenTaskVersions = new WeakMap<QueryClient, Map<string, number>>();

function versionsFor(queryClient: QueryClient): Map<string, number> {
  let versions = seenTaskVersions.get(queryClient);
  if (!versions) {
    versions = new Map();
    seenTaskVersions.set(queryClient, versions);
  }
  return versions;
}

interface ReconcileOptions {
  previousColumnId?: string;
  optimistic?: boolean;
}

export function reconcileTask(queryClient: QueryClient, task: Task, options: ReconcileOptions = {}) {
  const projectId = task.project_id;
  const targetColumnId = task.project_column_id;
  const current = findTaskOnBoard(queryClient, projectId, task.id);

  if (!options.optimistic) {
    const versions = versionsFor(queryClient);
    const lastSeen = Math.max(versions.get(task.id) ?? -1, current?.task.version ?? -1);
    if (task.version <= lastSeen) return;
    versions.set(task.id, task.version);
  }

  const invalidateCounts = () => {
    if (options.optimistic) return;
    queryClient.invalidateQueries({ queryKey: taskQueryKeys.counts(projectId) });
  };

  const invalidateArchived = () => {
    if (options.optimistic) return;
    queryClient.invalidateQueries({ queryKey: taskQueryKeys.archived(projectId) });
  };

  if (task.archived_at) {
    const archivedColumnId = current?.columnId ?? options.previousColumnId ?? targetColumnId;
    removeTask(queryClient, projectId, archivedColumnId, task.id);
    invalidateCounts();
    invalidateArchived();
    return;
  }

  const upsert = () =>
    upsertTask(queryClient, projectId, targetColumnId, task, { allowBeyondWindow: options.optimistic });

  if (!current) {
    upsert();
    invalidateCounts();
    invalidateArchived();
    return;
  }

  if (current.columnId === targetColumnId) {
    upsert();
    return;
  }

  removeTask(queryClient, projectId, current.columnId, task.id);
  upsert();
  invalidateCounts();
}

export function invalidateBoard(queryClient: QueryClient, projectId: string) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: taskQueryKeys.board(projectId) }),
    queryClient.invalidateQueries({ queryKey: taskQueryKeys.counts(projectId) }),
    queryClient.invalidateQueries({ queryKey: taskQueryKeys.archived(projectId) }),
  ]);
}

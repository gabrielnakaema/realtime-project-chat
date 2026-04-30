import type { InfiniteData, QueryClient } from '@tanstack/react-query';
import type { CursorPaginated } from '@/types/paginated';
import type { Task } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';
import { DEFAULT_TASK_LIMIT } from '@/constants/tasks';

type ColumnPage = Record<string, CursorPaginated<Task>>;
type ColumnCache = InfiniteData<ColumnPage>;

export function buildColumnQueryKey(projectId: string, projectColumnId: string) {
  return taskQueryKeys.listGroupedByProjectId({
    projectId,
    projectColumnIds: [projectColumnId],
    archived: false,
    limit: DEFAULT_TASK_LIMIT,
    taskOrder: '',
    updatedAt: null,
  });
}

export function updateTaskInColumn(queryClient: QueryClient, projectId: string, projectColumnId: string, task: Task) {
  queryClient.setQueryData<ColumnCache>(buildColumnQueryKey(projectId, projectColumnId), (old) => {
    if (!old) return old;
    return {
      ...old,
      pages: old.pages.map((page) => ({
        ...page,
        [projectColumnId]: {
          ...page[projectColumnId],
          data: page[projectColumnId].data.map((t) => (t.id === task.id ? task : t)),
        },
      })),
    };
  });
}

export function removeTaskFromColumn(queryClient: QueryClient, projectId: string, projectColumnId: string, taskId: string) {
  queryClient.setQueryData<ColumnCache>(buildColumnQueryKey(projectId, projectColumnId), (old) => {
    if (!old) return old;
    return {
      ...old,
      pages: old.pages.map((page) => ({
        ...page,
        [projectColumnId]: {
          ...page[projectColumnId],
          data: page[projectColumnId].data.filter((t) => t.id !== taskId),
        },
      })),
    };
  });
}

export function insertTaskAtCorrectPosition(
  queryClient: QueryClient,
  projectId: string,
  projectColumnId: string,
  task: Task,
) {
  queryClient.setQueryData<ColumnCache>(buildColumnQueryKey(projectId, projectColumnId), (old) => {
    if (!old) return old;

    const allTasks = old.pages.flatMap((page) => page[projectColumnId].data);
    const taskWasLoaded = allTasks.some((t) => t.id === task.id);
    const filteredTasks = allTasks.filter((t) => t.id !== task.id);

    const lastPage = old.pages[old.pages.length - 1];
    const hasMore = lastPage[projectColumnId].has_next;
    const lastOrder = filteredTasks.at(-1)?.order;
    const beyondRange = hasMore && lastOrder !== undefined && task.order > lastOrder;

    if (beyondRange && !taskWasLoaded) return old;

    let newAllTasks: Task[];
    if (beyondRange) {
      newAllTasks = filteredTasks;
    } else {
      const insertIdx = filteredTasks.findIndex((t) => t.order > task.order);
      newAllTasks =
        insertIdx === -1
          ? [...filteredTasks, task]
          : [...filteredTasks.slice(0, insertIdx), task, ...filteredTasks.slice(insertIdx)];
    }

    let offset = 0;
    const newPages = old.pages.map((page, i) => {
      const size = i < old.pages.length - 1 ? page[projectColumnId].data.length : newAllTasks.length - offset;
      const pageData = newAllTasks.slice(offset, offset + size);
      offset += size;
      return { ...page, [projectColumnId]: { ...page[projectColumnId], data: pageData } };
    });

    return { ...old, pages: newPages };
  });
}

export function adjustCountCache(
  queryClient: QueryClient,
  projectId: string,
  projectColumnIds: string[],
  delta: Partial<Record<string, number>>,
) {
  queryClient.setQueryData<Record<string, number>>(
    taskQueryKeys.countByColumn(projectId, projectColumnIds),
    (old) => {
      if (!old) return old;
      const next = { ...old };
      for (const [status, diff] of Object.entries(delta) as [string, number][]) {
        next[status] = Math.max(0, (next[status] || 0) + diff);
      }
      return next;
    },
  );
}

export function findTaskInColumnCaches(
  queryClient: QueryClient,
  projectId: string,
  projectColumnIds: string[],
  taskId: string,
): { task: Task; columnId: string } | undefined {
  for (const columnId of projectColumnIds) {
    const data = queryClient.getQueryData<ColumnCache>(buildColumnQueryKey(projectId, columnId));
    if (!data) continue;
    for (const page of data.pages) {
      const found = page[columnId].data.find((t) => t.id === taskId);
      if (found) return { task: found, columnId };
    }
  }
  return undefined;
}

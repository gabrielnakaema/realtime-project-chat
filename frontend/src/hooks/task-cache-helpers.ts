import type { InfiniteData, QueryClient } from '@tanstack/react-query';
import type { CursorPaginated } from '@/types/paginated';
import type { Task, TaskStatus } from '@/types/task';
import { taskQueryKeys } from '@/services/query-keys';
import { DEFAULT_TASK_LIMIT, TASK_STATUSES } from '@/constants/tasks';

type ColumnPage = Record<string, CursorPaginated<Task>>;
type ColumnCache = InfiniteData<ColumnPage>;

export function buildColumnQueryKey(projectId: string, status: TaskStatus) {
  return taskQueryKeys.listGroupedByProjectId({
    projectId,
    statuses: [status],
    limit: DEFAULT_TASK_LIMIT,
    taskOrder: '',
    updatedAt: null,
  });
}

export function updateTaskInColumn(queryClient: QueryClient, projectId: string, status: TaskStatus, task: Task) {
  queryClient.setQueryData<ColumnCache>(buildColumnQueryKey(projectId, status), (old) => {
    if (!old) return old;
    return {
      ...old,
      pages: old.pages.map((page) => ({
        ...page,
        [status]: {
          ...page[status],
          data: page[status].data.map((t) => (t.id === task.id ? task : t)),
        },
      })),
    };
  });
}

export function removeTaskFromColumn(queryClient: QueryClient, projectId: string, status: TaskStatus, taskId: string) {
  queryClient.setQueryData<ColumnCache>(buildColumnQueryKey(projectId, status), (old) => {
    if (!old) return old;
    return {
      ...old,
      pages: old.pages.map((page) => ({
        ...page,
        [status]: {
          ...page[status],
          data: page[status].data.filter((t) => t.id !== taskId),
        },
      })),
    };
  });
}

export function insertTaskAtCorrectPosition(
  queryClient: QueryClient,
  projectId: string,
  status: TaskStatus,
  task: Task,
) {
  queryClient.setQueryData<ColumnCache>(buildColumnQueryKey(projectId, status), (old) => {
    if (!old) return old;

    const allTasks = old.pages.flatMap((page) => page[status].data);
    const taskWasLoaded = allTasks.some((t) => t.id === task.id);
    const filteredTasks = allTasks.filter((t) => t.id !== task.id);

    const lastPage = old.pages[old.pages.length - 1];
    const hasMore = lastPage[status].has_next;
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
      const size = i < old.pages.length - 1 ? page[status].data.length : newAllTasks.length - offset;
      const pageData = newAllTasks.slice(offset, offset + size);
      offset += size;
      return { ...page, [status]: { ...page[status], data: pageData } };
    });

    return { ...old, pages: newPages };
  });
}

export function adjustCountCache(
  queryClient: QueryClient,
  projectId: string,
  delta: Partial<Record<TaskStatus, number>>,
) {
  queryClient.setQueryData<Record<TaskStatus, number>>(
    taskQueryKeys.countByStatus(projectId, [...TASK_STATUSES]),
    (old) => {
      if (!old) return old;
      const next = { ...old };
      for (const [status, diff] of Object.entries(delta) as [TaskStatus, number][]) {
        next[status] = Math.max(0, (next[status] || 0) + diff);
      }
      return next;
    },
  );
}

export function findTaskInColumnCaches(
  queryClient: QueryClient,
  projectId: string,
  taskId: string,
): { task: Task; status: TaskStatus } | undefined {
  for (const status of TASK_STATUSES) {
    const data = queryClient.getQueryData<ColumnCache>(buildColumnQueryKey(projectId, status));
    if (!data) continue;
    for (const page of data.pages) {
      const found = page[status].data.find((t) => t.id === taskId);
      if (found) return { task: found, status };
    }
  }
  return undefined;
}

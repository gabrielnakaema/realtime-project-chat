import type { InfiniteData, QueryClient } from '@tanstack/react-query';
import type { CursorPaginated } from '@/types/paginated';
import type { ListTasksRequest, Task } from '@/types/task';
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

export function removeTaskFromColumn(
  queryClient: QueryClient,
  projectId: string,
  projectColumnId: string,
  taskId: string,
) {
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

function applyCountDelta(old: Record<string, number> | undefined, delta: Partial<Record<string, number>>) {
  if (!old) return old;
  const next = { ...old };
  for (const [columnId, diff] of Object.entries(delta) as [string, number][]) {
    next[columnId] = Math.max(0, (next[columnId] || 0) + diff);
  }
  return next;
}

export function adjustBoardColumnCounts(
  queryClient: QueryClient,
  projectId: string,
  delta: Partial<Record<string, number>>,
) {
  const entries = queryClient.getQueriesData<Record<string, number>>({ queryKey: taskQueryKeys._allCounts() });
  for (const [key, data] of entries) {
    if (!data) continue;

    const params = key[2] as { projectId?: string } | undefined;
    if (!params || params.projectId !== projectId) continue;

    queryClient.setQueryData<Record<string, number>>(key, (old) => applyCountDelta(old, delta));
  }
}

export function findTaskInBoardCaches(
  queryClient: QueryClient,
  projectId: string,
  taskId: string,
): { task: Task; columnId: string } | undefined {
  const entries = queryClient.getQueriesData<ColumnCache>({ queryKey: taskQueryKeys._allGrouped() });
  for (const [key, data] of entries) {
    if (!data) continue;

    const request = key[3] as ListTasksRequest | undefined;
    const columnId = request?.projectColumnIds[0];
    if (!request || request.projectId !== projectId || request.archived || !columnId) continue;

    for (const page of data.pages) {
      const found = page[columnId].data.find((t) => t.id === taskId);
      if (found) return { task: found, columnId };
    }
  }
  return undefined;
}

export function reconcileTaskInBoard(
  queryClient: QueryClient,
  projectId: string,
  task: Task,
  options: { force?: boolean; previousColumnId?: string } = {},
) {
  const targetColumn = task.project_column_id;
  const current = findTaskInBoardCaches(queryClient, projectId, task.id);

  if (!options.force && current && current.task.version >= task.version) return;

  if (task.archived_at) {
    const archivedColumn = current?.columnId ?? options.previousColumnId ?? targetColumn;
    removeTaskFromColumn(queryClient, projectId, archivedColumn, task.id);
    adjustBoardColumnCounts(queryClient, projectId, { [archivedColumn]: -1 });
    return;
  }

  if (!current) {
    insertTaskAtCorrectPosition(queryClient, projectId, targetColumn, task);
    if (options.previousColumnId && options.previousColumnId !== targetColumn) {
      adjustBoardColumnCounts(queryClient, projectId, { [options.previousColumnId]: -1, [targetColumn]: 1 });
    } else {
      adjustBoardColumnCounts(queryClient, projectId, { [targetColumn]: 1 });
    }
    return;
  }

  if (current.columnId === targetColumn) {
    insertTaskAtCorrectPosition(queryClient, projectId, targetColumn, task);
    return;
  }

  removeTaskFromColumn(queryClient, projectId, current.columnId, task.id);
  insertTaskAtCorrectPosition(queryClient, projectId, targetColumn, task);
  adjustBoardColumnCounts(queryClient, projectId, { [current.columnId]: -1, [targetColumn]: 1 });
}

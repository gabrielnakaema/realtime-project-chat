import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import type { CursorPaginated } from '@/types/paginated';
import type { Project } from '@/types/project';
import type { ListTasksRequest, Task } from '@/types/task';
import { DEFAULT_TASK_LIMIT } from '@/constants/tasks';
import { useInfiniteScrollObserver } from '@/hooks/use-infinite-scroll-observer';
import { taskQueryKeys } from '@/services/query-keys';
import { listGroupedTasksByProjectId, restoreTask } from '@/services/tasks';

type ArchivedTasksPage = Record<string, CursorPaginated<Task>>;

interface ArchivedTasksPageParam {
  taskOrder: string;
  updatedAt: string | null;
}

interface RestoreArchivedTaskInput {
  projectColumnId: string;
  taskId: string;
}

const INITIAL_ARCHIVED_TASKS_PAGE_PARAM: ArchivedTasksPageParam = {
  taskOrder: '',
  updatedAt: null,
};

export function buildArchivedTasksRequest(
  projectId: string,
  pageParam: ArchivedTasksPageParam = INITIAL_ARCHIVED_TASKS_PAGE_PARAM,
): ListTasksRequest {
  return {
    projectId,
    projectColumnIds: [],
    archived: true,
    taskOrder: pageParam.taskOrder,
    updatedAt: pageParam.updatedAt,
    limit: DEFAULT_TASK_LIMIT,
  };
}

export function getNextArchivedTasksPageParam(lastPage: ArchivedTasksPage): ArchivedTasksPageParam | undefined {
  const firstColumn = Object.values(lastPage)[0];
  if (!firstColumn.has_next) {
    return undefined;
  }

  const lastTask = firstColumn.data.at(-1);
  if (!lastTask) {
    return undefined;
  }

  return {
    taskOrder: lastTask.order,
    updatedAt: lastTask.updated_at,
  };
}

export function flattenArchivedTasks(pages?: ArchivedTasksPage[]) {
  return pages?.flatMap((page) => Object.values(page).flatMap((column) => column.data)) ?? [];
}

export const useArchivedTasksList = (project: Project, open: boolean) => {
  const queryClient = useQueryClient();
  const [pickingStatusForTaskId, setPickingStatusForTaskId] = useState<string | null>(null);

  const query = useInfiniteQuery({
    queryKey: taskQueryKeys.listGroupedByProjectId(buildArchivedTasksRequest(project.id)),
    queryFn: ({ pageParam }) => listGroupedTasksByProjectId(buildArchivedTasksRequest(project.id, pageParam)),
    getNextPageParam: getNextArchivedTasksPageParam,
    initialPageParam: INITIAL_ARCHIVED_TASKS_PAGE_PARAM,
    enabled: open,
  });

  const archivedTasks = useMemo(() => flattenArchivedTasks(query.data?.pages), [query.data]);

  const {
    mutate: restore,
    isPending: isRestoring,
    variables: restoreVariables,
  } = useMutation({
    mutationFn: ({ taskId, projectColumnId }: RestoreArchivedTaskInput) => restoreTask(taskId, projectColumnId),
    onSuccess: () => {
      setPickingStatusForTaskId(null);
      queryClient.invalidateQueries({
        queryKey: taskQueryKeys.listGroupedByProjectId(buildArchivedTasksRequest(project.id)),
      });
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allGrouped() });
      queryClient.invalidateQueries({ queryKey: taskQueryKeys._allCounts() });
    },
  });

  const sentinelRef = useInfiniteScrollObserver<HTMLDivElement>({
    onLoadMore: () => {
      if (!query.hasNextPage || query.isFetchingNextPage) {
        return;
      }

      query.fetchNextPage();
    },
  });

  return {
    archivedTasks,
    isFetchingNextPage: query.isFetchingNextPage,
    pickingStatusForTaskId,
    sentinelRef,
    restoringProjectColumnId: isRestoring ? restoreVariables.projectColumnId : null,
    restoringTaskId: isRestoring ? restoreVariables.taskId : null,
    restoreTaskToColumn: (task: Task, projectColumnId: string) => restore({ taskId: task.id, projectColumnId }),
    startPickingStatus: (taskId: string) => setPickingStatusForTaskId(taskId),
    stopPickingStatus: () => setPickingStatusForTaskId(null),
  };
};

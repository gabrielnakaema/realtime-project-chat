import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import type { CursorPaginated } from '@/shared/types/paginated';
import type { Task } from '@/features/tasks/types/task';
import { DEFAULT_TASK_LIMIT } from '@/features/tasks/constants/tasks';
import { reconcileTask } from '@/features/tasks/hooks/task-board-cache';
import { useInfiniteScrollObserver } from '@/shared/hooks/use-infinite-scroll-observer';
import { taskQueryKeys } from '@/shared/services/query-keys';
import { listColumnTasks, restoreTask } from '@/features/tasks/services/tasks';

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

export function getNextArchivedTasksPageParam(lastPage: CursorPaginated<Task>): ArchivedTasksPageParam | undefined {
  if (!lastPage.has_next) {
    return undefined;
  }

  const lastTask = lastPage.data.at(-1);
  if (!lastTask) {
    return undefined;
  }

  return {
    taskOrder: lastTask.order,
    updatedAt: lastTask.updated_at,
  };
}

export const useArchivedTasksList = (projectId: string, open: boolean) => {
  const queryClient = useQueryClient();
  const [pickingStatusForTaskId, setPickingStatusForTaskId] = useState<string | null>(null);

  const query = useInfiniteQuery({
    queryKey: taskQueryKeys.archived(projectId),
    queryFn: ({ pageParam }) =>
      listColumnTasks({
        projectId: projectId,
        archived: true,
        limit: DEFAULT_TASK_LIMIT,
        taskOrder: pageParam.taskOrder,
        updatedAt: pageParam.updatedAt,
      }),
    getNextPageParam: getNextArchivedTasksPageParam,
    initialPageParam: INITIAL_ARCHIVED_TASKS_PAGE_PARAM,
    enabled: open,
  });

  const archivedTasks = useMemo(() => query.data?.pages.flatMap((page) => page.data) ?? [], [query.data]);

  const {
    mutate: restore,
    isPending: isRestoring,
    variables: restoreVariables,
  } = useMutation({
    mutationFn: ({ taskId, projectColumnId }: RestoreArchivedTaskInput) => restoreTask(taskId, projectColumnId),
    onSuccess: (restoredTask) => {
      setPickingStatusForTaskId(null);
      reconcileTask(queryClient, restoredTask);
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.archived(projectId) });
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

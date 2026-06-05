import { getRouteApi } from '@tanstack/react-router';
import { useState } from 'react';

const projectRouteApi = getRouteApi('/_protected/projects/$projectId/');

export const useTaskDetailsRouting = () => {
  const search = projectRouteApi.useSearch();
  const navigate = projectRouteApi.useNavigate();

  const { taskId, commentId, commentCreatedAt } = search;
  const [editingTaskId, setEditingTaskId] = useState<string | null>(null);

  const selectedTaskId = taskId ?? '';
  const selectedCommentId = commentId;
  const selectedCommentCreatedAt = commentCreatedAt;
  const isEditingTask = Boolean(selectedTaskId) && editingTaskId === selectedTaskId;

  const openTask = (nextTaskId: string) => {
    setEditingTaskId(null);
    navigate({
      search: (prev) => ({ ...prev, taskId: nextTaskId, commentId: undefined, commentCreatedAt: undefined }),
      replace: true,
    });
  };

  const closeTask = () => {
    setEditingTaskId(null);
    navigate({
      search: (prev) => ({ ...prev, taskId: undefined, commentId: undefined, commentCreatedAt: undefined }),
      replace: true,
    });
  };

  const startEditingTask = () => {
    if (!selectedTaskId) {
      return;
    }

    setEditingTaskId(selectedTaskId);
  };

  const stopEditingTask = () => {
    setEditingTaskId(null);
  };

  return {
    selectedTaskId,
    selectedCommentId,
    selectedCommentCreatedAt,
    isEditingTask,
    openTask,
    closeTask,
    startEditingTask,
    stopEditingTask,
  };
};

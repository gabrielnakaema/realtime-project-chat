import { useNavigate, useSearch } from '@tanstack/react-router';
import { useState } from 'react';

export const useTaskDetailsRouting = () => {
  const search = useSearch({
    from: '/projects/$projectId/',
  });
  const navigate = useNavigate({
    from: '/projects/$projectId',
  });

  const { taskId, commentId, commentCreatedAt } = search;
  const [isEditingTaskOpen, setIsEditingTaskOpen] = useState(false);

  const selectedTaskId = taskId ?? '';
  const selectedCommentId = commentId;
  const selectedCommentCreatedAt = commentCreatedAt;
  const isEditingTask = Boolean(selectedTaskId) && isEditingTaskOpen;

  const openTask = (nextTaskId: string) => {
    setIsEditingTaskOpen(false);
    navigate({
      search: (prev) => ({ ...prev, taskId: nextTaskId, commentId: undefined, commentCreatedAt: undefined }),
      replace: true,
    });
  };

  const closeTask = () => {
    setIsEditingTaskOpen(false);
    navigate({
      search: (prev) => ({ ...prev, taskId: undefined, commentId: undefined, commentCreatedAt: undefined }),
      replace: true,
    });
  };

  const startEditingTask = () => {
    if (!selectedTaskId) {
      return;
    }

    setIsEditingTaskOpen(true);
  };

  const stopEditingTask = () => {
    setIsEditingTaskOpen(false);
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

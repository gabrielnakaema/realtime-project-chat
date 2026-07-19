import type { Notification } from '@/features/notifications/types/notification';

export const getNotificationNavigationTarget = (notification: Notification) => {
  const projectId = notification.project?.id ?? notification.project_id;
  const taskId = notification.task?.id ?? notification.task_id;
  const commentId = notification.task_comment?.id ?? notification.task_comment_id;
  const commentCreatedAt = notification.task_comment?.created_at;

  return {
    to: '/projects/$projectId' as const,
    params: { projectId },
    search: taskId
      ? { taskId, commentId: commentId ?? undefined, commentCreatedAt: commentCreatedAt ?? undefined }
      : {},
  };
};

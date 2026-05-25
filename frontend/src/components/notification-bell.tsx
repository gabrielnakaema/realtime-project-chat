import { Bell, CheckCheck, Loader2 } from 'lucide-react';
import type { Notification } from '@/types/notification';
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useNotifications } from '@/hooks/use-notifications';
import { cn } from '@/lib/utils';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';

export const NotificationBell = () => {
  const {
    isOpen,
    setIsOpen,
    notifications,
    unreadCount,
    hasNextPage,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    openNotification,
    markAllAsRead,
  } = useNotifications();

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="relative rounded-md p-2 text-slate-700 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
          aria-label="Notifications"
        >
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <span className="absolute -top-1 -right-1 rounded-full bg-red-600 px-1.5 py-0.5 text-[10px] font-semibold text-white">
              {unreadCount > 99 ? '+99' : unreadCount}
            </span>
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[360px] p-0">
        <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3 dark:border-slate-700">
          <div>
            <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">Notifications</p>
            <p className="text-xs text-slate-500 dark:text-slate-400">Recent updates that need your attention</p>
          </div>
          <button
            type="button"
            onClick={markAllAsRead}
            disabled={unreadCount === 0}
            className="inline-flex items-center gap-1 text-xs font-medium text-slate-600 transition-colors hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-50 dark:text-slate-400 dark:hover:text-slate-100"
          >
            <CheckCheck className="h-3.5 w-3.5" />
            Mark all read
          </button>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center px-4 py-10">
            <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
          </div>
        ) : !notifications.length ? (
          <div className="px-4 py-10 text-center">
            <p className="text-sm font-medium text-slate-900 dark:text-slate-100">No notifications yet</p>
            <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">New assignments and comments will show up here.</p>
          </div>
        ) : (
          <>
            <ScrollArea className="h-[360px]">
              <div className="flex flex-col">
                {notifications.map((notification) => (
                  <NotificationRow key={notification.id} notification={notification} onClick={openNotification} />
                ))}
              </div>
            </ScrollArea>
            {hasNextPage && (
              <div className="border-t border-slate-200 p-2 dark:border-slate-700">
                <button
                  type="button"
                  onClick={() => fetchNextPage()}
                  disabled={isFetchingNextPage}
                  className="flex w-full items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900 disabled:cursor-not-allowed disabled:opacity-50 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-100"
                >
                  {isFetchingNextPage ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                  Load more
                </button>
              </div>
            )}
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

const NotificationRow = ({
  notification,
  onClick,
}: {
  notification: Notification;
  onClick: (notification: Notification) => Promise<void>;
}) => {
  return (
    <button
      type="button"
      onClick={() => onClick(notification)}
      className={cn(
        'flex w-full flex-col gap-2 border-b border-slate-200 px-4 py-3 text-left transition-colors hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800/70',
        !notification.read_at && 'bg-blue-50/70 dark:bg-slate-800',
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <p className="text-sm text-slate-900 dark:text-slate-100">{getNotificationText(notification)}</p>
        {!notification.read_at && <span className="mt-1 h-2 w-2 rounded-full bg-blue-600" aria-hidden="true" />}
      </div>
      <p className="text-xs text-slate-500 dark:text-slate-400">{formatRelativeActivityDateString(notification.created_at)}</p>
    </button>
  );
};

const getNotificationText = (notification: Notification) => {
  const actorName = notification.actor?.name ?? 'Someone';
  const projectName = notification.project?.name ?? 'a project';
  const taskTitle = notification.task?.title ?? 'a task';

  if (notification.type === 'project.member.created') {
    return `${actorName} added you to ${projectName}.`;
  }

  if (notification.type === 'task.assigned') {
    return `${actorName} assigned you to ${taskTitle}.`;
  }

  return `${actorName} commented on ${taskTitle}.`;
};

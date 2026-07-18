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
          className="text-foreground hover:bg-muted relative rounded-md p-2 transition-colors"
          aria-label="Notifications"
        >
          <Bell className="h-4 w-4" />
          {unreadCount > 0 && (
            <span className="bg-destructive text-primary-foreground absolute -top-1 -right-1 rounded-full px-1.5 py-0.5 text-[10px] font-semibold">
              {unreadCount > 99 ? '+99' : unreadCount}
            </span>
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[360px] p-0">
        <div className="border-border flex items-center justify-between border-b px-4 py-3">
          <div>
            <p className="text-foreground text-sm font-semibold">Notifications</p>
            <p className="text-muted-foreground text-xs">Recent updates that need your attention</p>
          </div>
          <button
            type="button"
            onClick={markAllAsRead}
            disabled={unreadCount === 0}
            className="text-muted-foreground hover:text-foreground inline-flex items-center gap-1 text-xs font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
          >
            <CheckCheck className="h-3.5 w-3.5" />
            Mark all read
          </button>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center px-4 py-10">
            <Loader2 className="text-muted-foreground h-4 w-4 animate-spin" />
          </div>
        ) : !notifications.length ? (
          <div className="px-4 py-10 text-center">
            <p className="text-foreground text-sm font-medium">No notifications yet</p>
            <p className="text-muted-foreground mt-1 text-xs">New assignments and comments will show up here.</p>
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
              <div className="border-border border-t p-2">
                <button
                  type="button"
                  onClick={() => fetchNextPage()}
                  disabled={isFetchingNextPage}
                  className="text-muted-foreground hover:bg-muted hover:text-foreground flex w-full items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
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
        'border-border hover:bg-muted flex w-full flex-col gap-2 border-b px-4 py-3 text-left transition-colors',
        !notification.read_at && 'bg-primary/10',
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <p className="text-foreground text-sm">{getNotificationText(notification)}</p>
        {!notification.read_at && <span className="bg-primary mt-1 h-2 w-2 rounded-full" aria-hidden="true" />}
      </div>
      <p className="text-muted-foreground text-xs">{formatRelativeActivityDateString(notification.created_at)}</p>
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

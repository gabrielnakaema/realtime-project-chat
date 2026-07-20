import { ChevronLeft, Cog, MessageSquare, Plus } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import type { Project } from '@/features/projects/types/project';
import { MembersAvatarList } from '@/shared/components/members-avatar-list';
import { NotificationBell } from '@/features/notifications/components/notification-bell';
import { ProjectsDropdown } from '@/features/projects/components/project-details-page/projects-dropdown';
import { ArchivedTasksModal } from '@/features/tasks/components/archived-tasks-modal';
import { CreateTask } from '@/features/tasks/components/task-form/create-task';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip';
import { UnreadCountBadge } from '@/features/chat/components/unread-count-badge';
import { useOnlineUsers } from '@/shared/hooks/use-online-users';
import { cn } from '@/lib/utils';
import { getChatByProjectId } from '@/features/chat/services/chat';
import { projectChatQueryKeys } from '@/shared/services/query-keys';
import { Button, buttonVariants } from '@/shared/components/button';
import { HeaderActions, HeaderLogo, HeaderShell, HeaderStart } from '@/shared/components/header-shell';

export const ProjectDetailsHeader = ({ project }: { project: Project }) => {
  const projectId = project.id;
  const initialProjectColumnId = [...project.columns].sort((left, right) => left.position - right.position)[0]?.id;
  const { onlineUserIds } = useOnlineUsers(projectId, 'project');

  const { data: chat } = useQuery({
    queryKey: projectChatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const unreadCount = chat?.unread_count ?? 0;
  const hasMoreUnread = chat?.has_more_unread ?? false;

  return (
    <HeaderShell>
      <HeaderStart>
        <div className="flex items-center gap-1">
          <HeaderLogo className="hidden sm:block" />

          <Link
            type="button"
            to="/projects"
            aria-label="Back to projects"
            className={cn(buttonVariants({ variant: 'ghost' }), 'size-9 px-0 sm:w-fit sm:px-4 sm:pl-1')}
          >
            <ChevronLeft className="h-4 w-4" />
            <h1 className="text-foreground hidden text-sm font-semibold tracking-[-0.01em] sm:block">TaskFlow</h1>
          </Link>
        </div>
        <ProjectsDropdown
          current={{
            label: project.name,
            value: project.id,
          }}
        />
      </HeaderStart>

      <HeaderActions>
        <CreateTask
          projectId={projectId}
          initialProjectColumnId={initialProjectColumnId}
          trigger={
            <Button type="button" className="size-9 px-0 md:w-fit md:px-4" aria-label="Create task">
              <Plus className="h-4 w-4" />
              <span className="hidden md:inline">Create task</span>
            </Button>
          }
        />
        <div className="hidden md:block">
          <ArchivedTasksModal projectId={projectId} />
        </div>
        <Tooltip>
          <TooltipTrigger asChild>
            <Link
              to="/projects/$projectId/chat"
              className={cn(buttonVariants({ variant: 'ghost', size: 'icon', className: 'relative' }))}
              params={{ projectId }}
              aria-label="Chat"
              title="Chat"
            >
              <MessageSquare className="h-4 w-4" />
              <UnreadCountBadge
                className="bg-primary absolute top-0 right-0 flex items-center justify-center p-1.5 py-0.5"
                count={unreadCount}
                hasMoreUnread={hasMoreUnread}
              />
            </Link>
          </TooltipTrigger>
          <TooltipContent>Chat</TooltipContent>
        </Tooltip>
        <div className="hidden lg:block">
          <NotificationBell />
        </div>

        <div className="hidden lg:block">
          <MembersAvatarList
            onlineUserIds={onlineUserIds}
            members={project.members.map((member) => ({ user_id: member.user_id, name: member.user.name }))}
            max={4}
          />
        </div>

        <Tooltip>
          <TooltipTrigger asChild>
            <Link
              to="/projects/$projectId/settings"
              className={cn(buttonVariants({ variant: 'ghost', size: 'icon', className: 'relative' }))}
              params={{ projectId }}
              title="Settings"
            >
              <Cog className="h-4 w-4" />
            </Link>
          </TooltipTrigger>
          <TooltipContent>Settings</TooltipContent>
        </Tooltip>
      </HeaderActions>
    </HeaderShell>
  );
};

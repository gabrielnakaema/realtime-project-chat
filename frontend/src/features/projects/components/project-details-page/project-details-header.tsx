import { ChevronLeft, Cog, MessageSquare, Plus } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { ProjectHeader } from '../project-header';
import type { Project } from '@/types/project';
import { MembersAvatarList } from '@/components/members-avatar-list';
import { NotificationBell } from '@/components/notification-bell';
import { ProjectsDropdown } from '@/features/projects/components/project-details-page/projects-dropdown';
import { ArchivedTasksModal } from '@/components/archived-tasks-modal';
import { CreateTask } from '@/components/task-form/create-task';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { UnreadCountBadge } from '@/components/unread-count-badge';
import { useOnlineUsers } from '@/hooks/use-online-users';
import { cn } from '@/lib/utils';
import { getChatByProjectId } from '@/services/chat';
import { projectChatQueryKeys } from '@/services/query-keys';
import { Button, buttonVariants } from '@/shared/components/button';

export const ProjectDetailsHeader = ({ project }: { project: Project }) => {
  const projectId = project.id;
  const { onlineUserIds } = useOnlineUsers(projectId, 'project');

  const { data: chat } = useQuery({
    queryKey: projectChatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const unreadCount = chat?.unread_count ?? 0;
  const hasMoreUnread = chat?.has_more_unread ?? false;

  return (
    <ProjectHeader>
      <div className="flex w-full items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1">
            <ProjectHeader.Logo />

            <Link type="button" to="/projects" className={cn(buttonVariants({ variant: 'ghost', className: 'pl-1' }))}>
              <ChevronLeft className="h-4 w-4" />
              <h1 className="text-foreground text-sm font-semibold tracking-[-0.01em]">Taskflow</h1>
            </Link>
          </div>
          <ProjectsDropdown
            current={{
              label: project.name,
              value: project.id,
            }}
          />
        </div>
      </div>

      <div className="flex items-center gap-4">
        <CreateTask
          projectId={projectId}
          trigger={
            <Button type="button">
              <Plus className="h-3 w-3" />
              Create task
            </Button>
          }
        />
        <ArchivedTasksModal projectId={projectId} />
        <Tooltip>
          <TooltipTrigger asChild>
            <Link
              to="/projects/$projectId/chat"
              className={cn(buttonVariants({ variant: 'ghost', size: 'icon', className: 'relative' }))}
              params={{ projectId }}
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
        <NotificationBell />

        <MembersAvatarList
          onlineUserIds={onlineUserIds}
          members={project.members.map((member) => ({ user_id: member.user_id, name: member.user.name }))}
          max={4}
        />

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
      </div>
    </ProjectHeader>
  );
};

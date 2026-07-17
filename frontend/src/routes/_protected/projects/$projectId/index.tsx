import { useQuery } from '@tanstack/react-query';
import { Link, createFileRoute } from '@tanstack/react-router';
import { ChevronLeft, Cog, MessageSquare, Plus } from 'lucide-react';
import { z } from 'zod';
import { ArchivedTasksModal } from '@/components/archived-tasks-modal';
import { KanbanBoard } from '@/components/kanban-board';
import { MembersAvatarList } from '@/components/members-avatar-list';
import { NotificationBell } from '@/components/notification-bell';
import { ProjectsDropdown } from '@/components/projects-page/projects-dropdown';
import { TaskDetails } from '@/components/task-details';
import { CreateTask } from '@/components/task-form/create-task';
import { EditTask } from '@/components/task-form/edit-task';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { UnreadCountBadge } from '@/components/unread-count-badge';
import { useOnlineUsers } from '@/hooks/use-online-users';
import { useProjectDetails } from '@/hooks/use-project-details';
import { useTaskDetailsRouting } from '@/hooks/use-task-details-routing';
import { getChatByProjectId } from '@/services/chat';
import { projectChatQueryKeys } from '@/services/query-keys';

export const Route = createFileRoute('/_protected/projects/$projectId/')({
  component: RouteComponent,
  validateSearch: z.object({
    taskId: z.string().optional(),
    commentId: z.string().optional(),
    commentCreatedAt: z.string().optional(),
  }),
});

function RouteComponent() {
  const { projectId } = Route.useParams();
  const {
    selectedTaskId,
    selectedCommentId,
    selectedCommentCreatedAt,
    isEditingTask,
    closeTask,
    startEditingTask,
    stopEditingTask,
  } = useTaskDetailsRouting();

  const { onlineUserIds } = useOnlineUsers(projectId, 'project');

  const { data: project } = useProjectDetails(projectId);

  const { data: chat } = useQuery({
    queryKey: projectChatQueryKeys.detailsByProjectId(projectId),
    queryFn: () => getChatByProjectId(projectId),
  });

  const unreadCount = chat?.unread_count ?? 0;
  const hasMoreUnread = chat?.has_more_unread ?? false;

  return (
    <div className="flex h-fit min-h-screen flex-col bg-slate-50 dark:bg-slate-900">
      <header className="flex min-h-16 w-full items-center border-b border-slate-200 bg-white px-4 dark:border-slate-700 dark:bg-slate-900">
        <div className="flex w-full items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1">
              <img src="/logo.png" className="h-8 w-8 object-contain" />

              <Link
                type="button"
                to="/projects"
                className="flex h-9 cursor-pointer items-center justify-center gap-2 rounded-md px-3 pl-1 hover:bg-slate-800 focus:bg-slate-800"
              >
                <ChevronLeft className="h-4 w-4" />
                <h1 className="text-foreground text-sm font-semibold tracking-[-0.01em]">Taskflow</h1>
              </Link>
            </div>
            <div className="h-4 min-w-[1px] dark:bg-slate-700" />
            <ProjectsDropdown
              current={{
                label: project?.name || '',
                value: project?.id || '',
              }}
            />
          </div>
        </div>

        <div className="flex items-center gap-4">
          <CreateTask
            projectId={projectId}
            trigger={
              <button
                type="button"
                className="flex min-h-9 items-center gap-2 rounded-md border border-blue-800 bg-blue-500/10 px-3 py-1 text-sm font-semibold whitespace-nowrap transition-colors hover:bg-blue-500/40"
              >
                <Plus className="h-3 w-3" />
                Create task
              </button>
            }
          />
          <ArchivedTasksModal projectId={projectId} />
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                to="/projects/$projectId/chat"
                className="relative inline-flex h-9 items-center rounded-md px-3 py-2 font-medium text-slate-700 transition-colors hover:bg-slate-50 dark:text-slate-300 dark:hover:bg-slate-600"
                params={{ projectId }}
                title="Chat"
              >
                <MessageSquare className="h-4 w-4" />
                <UnreadCountBadge
                  className="absolute top-0 right-0 flex items-center justify-center bg-blue-700 p-1.5 py-0.5"
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
            members={project?.members.map((member) => ({ user_id: member.user_id, name: member.user.name })) || []}
            max={4}
          />

          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                to="/projects/$projectId/settings"
                className="relative inline-flex h-9 items-center rounded-md px-3 py-2 font-medium text-slate-700 transition-colors hover:bg-slate-50 dark:text-slate-300 dark:hover:bg-slate-600"
                params={{ projectId }}
                title="Settings"
              >
                <Cog className="h-4 w-4" />
              </Link>
            </TooltipTrigger>
            <TooltipContent>Settings</TooltipContent>
          </Tooltip>
        </div>
      </header>

      {project && (
        <div className="flex-1 bg-slate-950 p-6">
          <KanbanBoard project={project} />
        </div>
      )}

      <TaskDetails
        taskId={selectedTaskId}
        open={!!selectedTaskId && !isEditingTask}
        targetCommentId={selectedCommentId}
        targetCommentCreatedAt={selectedCommentCreatedAt}
        onOpenChange={(open) => {
          if (!open) {
            closeTask();
          }
        }}
        onEdit={startEditingTask}
      />

      <EditTask onOpenChange={stopEditingTask} taskId={selectedTaskId} open={isEditingTask} />
    </div>
  );
}

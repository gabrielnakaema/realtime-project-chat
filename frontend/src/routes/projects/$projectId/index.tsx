import { useQuery } from '@tanstack/react-query';
import { Link, createFileRoute } from '@tanstack/react-router';
import { ArrowLeft, MessageSquare } from 'lucide-react';
import { AddProjectMember } from '@/components/add-project-member';
import { KanbanBoard } from '@/components/kanban-board';
import { MembersAvatarList } from '@/components/members-avatar-list';
import { ProjectDetailsSheet, ProjectDetailsSheetTrigger } from '@/components/project-details-sheet';
import { ProjectSettings } from '@/components/project-settings';
import { useOnlineUsers } from '@/hooks/use-online-users';
import { getProject } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';
import { sanitizeHTML } from '@/utils/html';
import { ProjectMembersModal } from '@/components/project-members-modal';

export const Route = createFileRoute('/projects/$projectId/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { projectId } = Route.useParams();

  const { onlineUserIds } = useOnlineUsers(projectId);

  const { data: project } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
  });

  return (
    <div className="h-fit min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <div className="px-6 py-4">
          <div className="flex flex-col items-center justify-between gap-4 md:flex-row">
            <div className="flex items-center gap-4">
              <Link
                to="/projects"
                className="inline-flex items-center rounded-md px-3 py-2 font-medium whitespace-nowrap text-slate-700 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Go back
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">{project?.name}</h1>
                {project && (
                  <ProjectDetailsSheet project={project}>
                    <div className="flex items-center gap-1">
                      {project.description && (
                        <p
                          className="line-clamp-1 text-sm text-slate-600 dark:text-slate-400"
                          dangerouslySetInnerHTML={{ __html: sanitizeHTML(project.description) }}
                        />
                      )}
                      <ProjectDetailsSheetTrigger asChild>
                        <button
                          type="button"
                          className="shrink-0 text-sm font-medium text-blue-600 transition-colors hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
                        >
                          View details
                        </button>
                      </ProjectDetailsSheetTrigger>
                    </div>
                  </ProjectDetailsSheet>
                )}
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <AddProjectMember projectId={projectId} />
                <ProjectMembersModal project={project} />

                <MembersAvatarList
                  onlineUserIds={onlineUserIds}
                  members={
                    project?.members.map((member) => ({ user_id: member.user_id, name: member.user.name })) || []
                  }
                  max={4}
                />
              </div>
              <Link
                to="/projects/$projectId/chat"
                className="inline-flex items-center rounded-md border border-slate-300 bg-white px-3 py-2 font-medium text-slate-700 transition-colors hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-300 dark:hover:bg-slate-600"
                params={{ projectId }}
              >
                <MessageSquare className="mr-2 h-4 w-4" />
                Chat
              </Link>
              <ProjectSettings projectId={projectId} />
            </div>
          </div>
        </div>
      </header>

      {project && (
        <div className="p-6">
          <KanbanBoard project={project} />
        </div>
      )}
    </div>
  );
}

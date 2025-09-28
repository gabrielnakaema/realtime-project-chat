import { AddProjectMember } from '@/components/add-project-member';
import { KanbanBoard } from '@/components/kanban-board';
import { MembersAvatarList } from '@/components/members-avatar-list';
import { useOnlineUsers } from '@/hooks/use-online-users';
import { getProject } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute, Link } from '@tanstack/react-router';
import { ArrowLeft, MessageSquare, Settings, Users } from 'lucide-react';

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
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 ">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between md:flex-row flex-col gap-4">
            <div className="flex items-center gap-4">
              <Link
                to="/projects"
                className="inline-flex items-center px-3 py-2 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-md font-medium transition-colors"
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Go back
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">{project?.name}</h1>
                <p className="text-slate-600 dark:text-slate-400">{project?.description}</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <AddProjectMember projectId={projectId} />
                <Users className="w-4 h-4 text-slate-500" />

                <MembersAvatarList
                  onlineUserIds={onlineUserIds}
                  members={
                    project?.members?.map((member) => ({ user_id: member.user_id, name: member.user.name })) || []
                  }
                  max={4}
                />
              </div>
              <Link
                to="/projects/$projectId/chat"
                className="inline-flex items-center px-3 py-2 text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-600 rounded-md font-medium transition-colors"
                params={{ projectId }}
              >
                <MessageSquare className="w-4 h-4 mr-2" />
                Chat
              </Link>
              <button className="inline-flex items-center px-3 py-2 text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-600 rounded-md font-medium transition-colors">
                <Settings className="w-4 h-4 mr-2" />
                Settings
              </button>
            </div>
          </div>
        </div>
      </header>

      {project && (
        <div className="p-6">
          <KanbanBoard projectId={project.id} />
        </div>
      )}
    </div>
  );
}

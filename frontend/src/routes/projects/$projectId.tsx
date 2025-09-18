import { AddProjectMember } from '@/components/add-project-member';
import { KanbanBoard } from '@/components/kanban-board';
import { getProject } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute, Link } from '@tanstack/react-router';
import { ArrowLeft, Settings, Users } from 'lucide-react';

export const Route = createFileRoute('/projects/$projectId')({
  component: RouteComponent,
});

function RouteComponent() {
  const { projectId } = Route.useParams();

  const { data: project } = useQuery({
    queryKey: projectQueryKeys.details(projectId),
    queryFn: () => getProject(projectId),
  });

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
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
                <div className="flex -space-x-2">
                  {project?.members.slice(0, 4).map((member) => (
                    <div
                      key={member.id}
                      className="w-8 h-8 bg-blue-600 rounded-full border-2 border-white dark:border-slate-800 flex items-center justify-center text-white text-xs font-medium"
                    >
                      {member?.user?.name.charAt(0).toUpperCase()}
                    </div>
                  ))}
                  {(project?.members?.length || 0) > 4 && (
                    <div className="w-8 h-8 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-white dark:border-slate-800 flex items-center justify-center">
                      <span className="text-xs font-medium text-slate-600 dark:text-slate-400">
                        +{(project?.members?.length || 0) - 4}
                      </span>
                    </div>
                  )}
                </div>
              </div>
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

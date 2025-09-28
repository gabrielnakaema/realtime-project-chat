import { useQuery } from '@tanstack/react-query';
import { Link, createFileRoute } from '@tanstack/react-router';
import { LogOut, Plus } from 'lucide-react';
import { Button } from '@/components/button';
import { CreateProject } from '@/components/create-project';
import { useAuth } from '@/hooks/use-auth';
import { listProjects } from '@/services/projects';
import { projectQueryKeys } from '@/services/query-keys';

export const Route = createFileRoute('/projects/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { user, logout } = useAuth();

  const handleLogout = () => {
    logout();
  };

  const { data: projects } = useQuery({
    queryKey: projectQueryKeys.list,
    queryFn: listProjects,
  });

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">TaskFlow</h1>
              <p className="text-slate-600 dark:text-slate-400">Welcome back, {user?.name}</p>
            </div>
            <div className="flex items-center gap-4">
              <CreateProject />
              <Button onClick={handleLogout} variant="secondary">
                <LogOut className="w-4 h-4" />
                Logout
              </Button>
              <div className="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center text-white font-medium">
                {user?.name.charAt(0).toUpperCase()}
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="px-6 py-12">
        <div className="mb-8">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Your Projects</h2>
          </div>
          {projects?.length === 0 ? (
            <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-12 text-center">
              <div className="w-16 h-16 bg-slate-100 dark:bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-4">
                <Plus className="w-8 h-8 text-slate-400" />
              </div>
              <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2">No projects yet</h3>
              <p className="text-slate-600 dark:text-slate-400 mb-6">
                Create your first project to start organizing tasks and collaborating with your team.
              </p>
              <div className="flex justify-center">
                <CreateProject />
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {projects?.map((project) => (
                <div
                  key={project.id}
                  className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition-shadow cursor-pointer"
                >
                  <Link to="/projects/$projectId" params={{ projectId: project.id }}>
                    <div className="p-6">
                      <div className="flex items-center justify-between mb-4">
                        <div className="w-3 h-3 rounded-full" />
                        <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-full">
                          {project.members.length || 0} members
                        </span>
                      </div>
                      <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2">{project.name}</h3>
                      <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">{project.description}</p>
                      <div className="space-y-3">
                        <div className="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
                          <span>
                            Last activity:{' '}
                            {project.updated_at ? new Date(project.updated_at).toLocaleDateString() : 'N/A'}
                          </span>
                        </div>
                      </div>
                    </div>
                  </Link>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

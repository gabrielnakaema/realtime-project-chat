import { createFileRoute } from '@tanstack/react-router';
import { LogOut } from 'lucide-react';
import { Button } from '@/components/button';
import { CreateProject } from '@/components/create-project';
import { ProjectList } from '@/components/project-list';
import { UserProjectActivities } from '@/components/user-project-activities';
import { useAuth } from '@/hooks/use-auth';

export const Route = createFileRoute('/projects/')({
  component: RouteComponent,
});

function RouteComponent() {
  const { user, logout } = useAuth();

  const handleLogout = () => {
    logout();
  };

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">TaskFlow</h1>
              <p className="text-slate-600 dark:text-slate-400">Welcome back, {user?.name}</p>
            </div>
            <div className="flex items-center gap-4">
              <CreateProject />
              <Button onClick={handleLogout} variant="secondary">
                <LogOut className="h-4 w-4" />
                Logout
              </Button>
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-600 font-medium text-white">
                {user?.name.charAt(0).toUpperCase()}
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="space-y-8 px-6 py-12">
        <ProjectList />

        <UserProjectActivities />
      </div>
    </div>
  );
}

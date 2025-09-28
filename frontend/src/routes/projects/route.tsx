import { Navigate, Outlet, createFileRoute } from '@tanstack/react-router';
import { useAuth } from '@/hooks/use-auth';

export const Route = createFileRoute('/projects')({
  component: RouteComponent,
});

function RouteComponent() {
  const { authStatus } = useAuth();

  if (authStatus === 'unauthenticated') {
    return <Navigate to="/login" />;
  }

  return <Outlet />;
}

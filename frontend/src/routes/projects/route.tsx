import { useAuth } from '@/hooks/use-auth';
import { createFileRoute, Navigate, Outlet } from '@tanstack/react-router';

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

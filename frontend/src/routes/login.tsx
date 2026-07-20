import { Link, Navigate, createFileRoute } from '@tanstack/react-router';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useMutation } from '@tanstack/react-query';
import type { SubmitHandler } from 'react-hook-form';
import type { ILoginForm } from '@/features/auth/schemas/login.schema';
import { Input } from '@/components/input';
import { loginSchema } from '@/features/auth/schemas/login.schema';
import { login } from '@/features/auth/services/auth';
import { LoadingSpinner } from '@/shared/components/loading';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { Button } from '@/components/button';

export const Route = createFileRoute('/login')({
  component: RouteComponent,
});

function RouteComponent() {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ILoginForm>({
    resolver: zodResolver(loginSchema),
  });
  const { authenticate, authStatus } = useAuth();

  const { mutate, isPending } = useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      authenticate(data);
    },
  });

  const onSubmit: SubmitHandler<ILoginForm> = (form) => {
    mutate(form);
  };

  if (authStatus === 'authenticated') {
    return <Navigate to="/projects" />;
  }

  return (
    <div className="from-muted to-muted flex min-h-screen items-center justify-center bg-gradient-to-br p-4">
      <main className="border-border bg-card w-full max-w-md rounded-lg border p-6">
        <div className="mb-6 text-center">
          <h1 className="text-foreground mb-2 text-2xl font-bold">Welcome back</h1>
          <p className="text-muted-foreground">Sign in to your account</p>
        </div>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            id="email"
            type="email"
            placeholder="Enter your email"
            label="Email"
            error={errors.email?.message}
            {...register('email')}
          />
          <Input
            id="password"
            type="password"
            placeholder="Enter your password"
            label="Password"
            error={errors.password?.message}
            {...register('password')}
          />

          <Button type="submit" disabled={isPending} className="w-full">
            {isPending ? <LoadingSpinner size="1.5em" /> : 'Sign In'}
          </Button>
        </form>
        <div className="mt-6 text-center">
          <p className="text-muted-foreground text-sm">
            Don't have an account?{' '}
            <Link to="/sign-up" className="text-primary hover:text-primary/80 font-medium">
              Sign up
            </Link>
          </p>
        </div>
      </main>
    </div>
  );
}

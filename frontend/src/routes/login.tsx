import { Link, Navigate, createFileRoute } from '@tanstack/react-router';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useMutation } from '@tanstack/react-query';
import type { SubmitHandler } from 'react-hook-form';
import type { ILoginForm } from '@/schemas/login-schema';
import { Input } from '@/components/input';
import { loginSchema } from '@/schemas/login-schema';
import { login } from '@/services/auth';
import { LoadingSpinner } from '@/components/loading';
import { useAuth } from '@/hooks/use-auth';
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
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 p-4 dark:from-slate-900 dark:to-slate-800">
      <main className="w-full max-w-md rounded-lg border border-slate-200 bg-white p-6 dark:border-slate-700 dark:bg-slate-800">
        <div className="mb-6 text-center">
          <h1 className="mb-2 text-2xl font-bold text-slate-900 dark:text-slate-100">Welcome back</h1>
          <p className="text-slate-600 dark:text-slate-400">Sign in to your account</p>
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
          <p className="text-sm text-slate-600 dark:text-slate-400">
            Don't have an account?{' '}
            <Link to="/sign-up" className="font-medium text-blue-600 hover:text-blue-700">
              Sign up
            </Link>
          </p>
        </div>
      </main>
    </div>
  );
}

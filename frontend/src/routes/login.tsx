import { loginSchema, type ILoginForm } from '@/schemas/login.schema';
import { createFileRoute, Link } from '@tanstack/react-router';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm, type SubmitHandler } from 'react-hook-form';
import { Input } from '@/components/input';
import { useMutation } from '@tanstack/react-query';
import { login } from '@/services/auth';
import { LoadingSpinner } from '@/components/loading';
import { useAuth } from '@/hooks/use-auth';

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
  const { authenticate } = useAuth();

  const { mutate, isPending } = useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      authenticate(data);
    },
  });

  const onSubmit: SubmitHandler<ILoginForm> = (form) => {
    mutate(form);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-6">
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">Welcome back</h1>
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

          <button
            type="submit"
            disabled={isPending}
            className="w-full px-4 py-2 text-white bg-blue-600 cursor-pointer hover:bg-blue-700 disabled:bg-blue-400 rounded-md font-medium transition-colors"
          >
            {isPending ? <LoadingSpinner size="1.5em" /> : 'Sign In'}
          </button>
        </form>
        <div className="mt-6 text-center">
          <p className="text-sm text-slate-600 dark:text-slate-400">
            Don't have an account?{' '}
            <Link to="/" className="text-blue-600 hover:text-blue-700 font-medium">
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}

import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
import type { SubmitHandler } from 'react-hook-form';
import type { ISignUpForm } from '@/features/auth/schemas/sign-up.schema';
import { Button } from '@/shared/components/button';
import { Input } from '@/shared/components/input';
import { LoadingSpinner } from '@/shared/components/loading';
import { signUpSchema } from '@/features/auth/schemas/sign-up.schema';
import { createUser } from '@/features/auth/services/users';
import { handleSuccess } from '@/shared/utils/handle-success';

export const Route = createFileRoute('/sign-up')({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ISignUpForm>({
    resolver: zodResolver(signUpSchema),
  });

  const { mutate, isPending } = useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      handleSuccess('User created successfully');
      navigate({ to: '/login' });
    },
  });

  const onSubmit: SubmitHandler<ISignUpForm> = (form) => {
    mutate(form);
  };

  return (
    <div className="from-muted to-muted flex min-h-screen items-center justify-center bg-gradient-to-br p-4">
      <div className="border-border bg-card w-full max-w-md rounded-lg border p-6">
        <div className="mb-6 text-center">
          <h1 className="text-foreground mb-2 text-2xl font-bold">Create Account</h1>
          <p className="text-muted-foreground">Join TaskFlow and start organizing your projects</p>
        </div>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            id="name"
            label="Name"
            placeholder="Enter your name"
            error={errors.name?.message}
            {...register('name')}
          />
          <Input
            id="email"
            label="Email"
            placeholder="Enter your email"
            error={errors.email?.message}
            {...register('email')}
          />
          <Input
            id="password"
            type="password"
            label="Password"
            placeholder="Create a password"
            error={errors.password?.message}
            {...register('password')}
          />
          <Input
            id="confirmPassword"
            type="password"
            label="Confirm Password"
            placeholder="Confirm your password"
            error={errors.confirmPassword?.message}
            {...register('confirmPassword')}
          />
          <Button type="submit" disabled={isPending} className="w-full">
            {isPending ? <LoadingSpinner size="1.5em" /> : 'Sign up'}
          </Button>
        </form>
        <div className="mt-6 text-center">
          <p className="text-muted-foreground text-sm">
            Already have an account?{' '}
            <Link to="/login" className="text-primary hover:text-primary/80 font-medium">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}

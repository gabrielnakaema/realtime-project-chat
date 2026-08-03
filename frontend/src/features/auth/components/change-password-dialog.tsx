import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { KeyRound } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../shared/components/ui/dialog';
import type { IChangePasswordForm } from '@/features/auth/schemas/change-password.schema';
import { Button } from '@/shared/components/button';
import { Input } from '@/shared/components/input';
import { LoadingSpinner } from '@/shared/components/loading';
import { changePasswordSchema } from '@/features/auth/schemas/change-password.schema';
import { changePassword } from '@/features/auth/services/users';
import { getErrorMessage } from '@/shared/utils/handle-error';
import { handleSuccess } from '@/shared/utils/handle-success';

interface ChangePasswordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const defaultValues: IChangePasswordForm = {
  old_password: '',
  new_password: '',
  new_password_confirmation: '',
};

export const ChangePasswordDialog = ({ open, onOpenChange }: ChangePasswordDialogProps) => {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<IChangePasswordForm>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues,
  });

  const closeDialog = () => {
    reset(defaultValues);
    setErrorMessage(null);
    onOpenChange(false);
  };

  const handleDialogOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      closeDialog();
      return;
    }

    onOpenChange(true);
  };

  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const changePasswordMutation = useMutation({
    mutationFn: changePassword,
    onSuccess: () => {
      handleSuccess('Password updated');
      closeDialog();
    },
    onError: async (error) => {
      setErrorMessage(await getErrorMessage(error));
    },
  });

  return (
    <Dialog open={open} onOpenChange={handleDialogOpenChange}>
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] flex-col gap-0 overflow-hidden p-0 sm:max-w-lg">
        <DialogHeader className="border-border bg-muted/80 border-b px-6 py-5">
          <DialogTitle className="flex items-center gap-2">
            <KeyRound className="text-primary h-5 w-5" />
            Change password
          </DialogTitle>
          <DialogDescription>Update your password without leaving the current page.</DialogDescription>
        </DialogHeader>

        <form
          onSubmit={handleSubmit((values) => changePasswordMutation.mutate(values))}
          className="flex flex-1 flex-col"
        >
          <div className="flex-1 space-y-4 overflow-y-auto px-6 py-5">
            <Input
              id="old_password"
              type="password"
              label="Current password"
              placeholder="Enter your current password"
              error={errors.old_password?.message}
              {...register('old_password')}
            />
            <Input
              id="new_password"
              type="password"
              label="New password"
              placeholder="Enter your new password"
              error={errors.new_password?.message}
              {...register('new_password')}
            />
            <Input
              id="new_password_confirmation"
              type="password"
              label="Confirm new password"
              placeholder="Confirm your new password"
              error={errors.new_password_confirmation?.message}
              {...register('new_password_confirmation')}
            />

            {errorMessage && <p className="text-destructive text-sm">{errorMessage}</p>}
          </div>

          <DialogFooter className="border-border bg-muted/80 border-t px-6 py-4">
            <Button type="button" variant="outline" onClick={closeDialog}>
              Cancel
            </Button>
            <Button type="submit" disabled={changePasswordMutation.isPending}>
              {changePasswordMutation.isPending ? <LoadingSpinner size="1.25rem" /> : 'Update password'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

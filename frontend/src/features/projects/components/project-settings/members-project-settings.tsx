import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import type { AddMemberFormData } from '@/features/projects/schemas/add-member.schema';
import type { Member } from '@/features/projects/types/project';
import type { SubmitHandler } from 'react-hook-form';
import { LoadingSpinner } from '@/shared/components/loading';
import { addMemberSchema } from '@/features/projects/schemas/add-member.schema';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { useProjectDetails } from '@/features/projects/hooks/use-project-details';
import { invalidateProjectBoardData } from '@/features/projects/services/project-board-invalidation';
import { createProjectMember, removeProjectMember } from '@/features/projects/services/projects';
import { projectQueryKeys } from '@/shared/services/query-keys';
import { Button } from '@/shared/components/button';
import { ConfirmationDialog } from '@/shared/components/confirmation-dialog';
import { Input } from '@/shared/components/input';
import { ProjectMemberRole } from '@/features/projects/types/project';
import { handleError } from '@/shared/utils/handle-error';
import { handleSuccess } from '@/shared/utils/handle-success';

const getInitials = (name: string) => {
  const [first = '', last = ''] = name.trim().split(' ');

  const firstInitial = first.charAt(0).trim();
  const lastInitial = last.charAt(0).trim();

  return `${firstInitial || ''}${lastInitial || ''}`;
};

export const MembersProjectSettings = ({ projectId }: { projectId: string }) => {
  const { user } = useAuth();
  const userId = user?.id;
  const { data: project, isLoading } = useProjectDetails(projectId);
  const [removingMember, setRemovingMember] = useState<Member | null>(null);
  const isCreator = project?.members.some(
    (member) => member.user_id === userId && member.role === ProjectMemberRole.Creator,
  );
  const queryClient = useQueryClient();
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AddMemberFormData>({
    resolver: zodResolver(addMemberSchema),
    defaultValues: {
      email: '',
    },
  });

  const removeMemberMutation = useMutation({
    mutationFn: removeProjectMember,
    onSuccess: () => {
      setRemovingMember(null);
      void invalidateProjectBoardData(queryClient);
    },
    onError: (error) => {
      void handleError(error);
    },
  });

  const addMemberMutation = useMutation({
    mutationFn: createProjectMember,
    onSuccess: () => {
      handleSuccess('Member added successfully');
      reset();
      void queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
    },
    onError: (error) => {
      void handleError(error);
    },
  });

  const handleRemoveMember = (member: Member) => {
    setRemovingMember(member);
  };

  const confirmRemoveMember = () => {
    if (!removingMember || removeMemberMutation.isPending) return;

    removeMemberMutation.mutate({
      projectId,
      memberUserId: removingMember.user_id,
    });
  };

  const handleConfirmationOpenChange = (open: boolean) => {
    if (!open && !removeMemberMutation.isPending) {
      setRemovingMember(null);
    }
  };

  const handleAddMember: SubmitHandler<AddMemberFormData> = (form) => {
    addMemberMutation.mutate({
      projectId,
      email: form.email,
    });
  };

  if (isLoading) {
    return <MembersProjectSettingsSkeleton />;
  }

  return (
    <div className="flex w-full flex-col gap-4">
      <section className="flex w-full flex-col gap-4">
        <p className="text-foreground text-xs font-semibold tracking-wider">
          MEMBERS &bull; {project?.members.length || 0}
        </p>
        <div className="w-full">
          {project?.members.map((member) => (
            <article key={member.id} className="border-border flex w-full items-center justify-between border-b py-4">
              <div className="flex items-center gap-2">
                <div className="border-border bg-card text-foreground flex h-8 min-h-8 w-8 min-w-8 items-center justify-center rounded-full border text-[11px] font-semibold uppercase">
                  {getInitials(member.user.name || '')}
                </div>
                <div>
                  <p className="text-foreground text-[13px] font-semibold">{member.user.name}</p>
                  <p className="text-muted-foreground text-xs font-normal">{member.user.email}</p>
                </div>
              </div>

              <div className="flex items-center gap-4">
                <p className="text-muted-foreground text-xs capitalize">{member.role}</p>

                {isCreator && member.role !== ProjectMemberRole.Creator && (
                  <Button
                    type="button"
                    variant="destructiveOutline"
                    size="sm"
                    aria-label={`Remove ${member.user.name} from project`}
                    onClick={() => handleRemoveMember(member)}
                  >
                    Remove
                  </Button>
                )}
              </div>
            </article>
          ))}
        </div>
      </section>

      {isCreator && (
        <section className="flex flex-col gap-4">
          <p className="text-foreground text-xs font-semibold tracking-wider">INVITE SOMEONE</p>
          <form className="flex items-start gap-2" noValidate onSubmit={handleSubmit(handleAddMember)}>
            <Input
              id="member-email"
              type="email"
              label="Email address"
              placeholder="Enter email address"
              error={errors.email?.message}
              classNames={{
                label: 'sr-only',
                error: 'text-destructive text-xs',
              }}
              {...register('email')}
            />
            <Button type="submit" className="h-[39px]" disabled={addMemberMutation.isPending}>
              {addMemberMutation.isPending ? <LoadingSpinner size="1.25rem" /> : 'Add member'}
            </Button>
          </form>
        </section>
      )}

      <ConfirmationDialog
        open={removingMember !== null}
        onOpenChange={handleConfirmationOpenChange}
        onConfirm={confirmRemoveMember}
        title="Remove member from project?"
        description={
          removingMember ? `Are you sure you want to remove ${removingMember.user.name} from the project?` : undefined
        }
        confirmLabel={removeMemberMutation.isPending ? 'Removing...' : 'Remove member'}
        variant="destructive"
        isConfirming={removeMemberMutation.isPending}
      />
    </div>
  );
};

const MembersProjectSettingsSkeleton = () => {
  return (
    <div className="flex w-full animate-pulse flex-col gap-4" aria-label="Loading project members">
      <div className="bg-card h-3 w-24 rounded" />
      {Array.from({ length: 3 }).map((_, index) => (
        <div key={index} className="border-border flex items-center justify-between border-b py-4">
          <div className="flex items-center gap-2">
            <div className="bg-card size-8 rounded-full" />
            <div className="flex flex-col gap-1.5">
              <div className="bg-card h-3 w-28 rounded" />
              <div className="bg-card h-3 w-40 rounded" />
            </div>
          </div>
          <div className="bg-card h-8 w-16 rounded-md" />
        </div>
      ))}
    </div>
  );
};

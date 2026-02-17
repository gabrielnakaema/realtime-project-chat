import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ShieldUser, Trash, User, Users } from 'lucide-react';
import { useState } from 'react';
import { Button } from './button';
import { LoadingSpinner } from './loading';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog';
import type { Member, Project } from '@/types/project';
import { ProjectMemberRole } from '@/types/project';
import { projectQueryKeys, taskQueryKeys } from '@/services/query-keys';
import { removeProjectMember } from '@/services/projects';
import { cn } from '@/lib/utils';
import { useAuth } from '@/hooks/use-auth';

interface ProjectMembersModalProps {
  project?: Project;
}

const memberRoleLabel: Record<string, string> = {
  [ProjectMemberRole.Creator]: 'Project creator',
  [ProjectMemberRole.Member]: 'Member',
};

const memberRoleIcon: Record<string, React.ReactNode> = {
  [ProjectMemberRole.Creator]: <ShieldUser className="h-4 w-4" />,
  [ProjectMemberRole.Member]: <User className="h-4 w-4" />,
};

export const ProjectMembersModal = ({ project }: ProjectMembersModalProps) => {
  const [open, setOpen] = useState(false);
  const { user } = useAuth();
  const userId = user?.id;
  const isCreator = project?.members.some(
    (member) => member.user_id === userId && member.role === ProjectMemberRole.Creator,
  );
  const [removingMember, setRemovingMember] = useState<Member | null>(null);
  const queryClient = useQueryClient();

  const { mutate, isPending } = useMutation({
    mutationFn: removeProjectMember,
    onSuccess: () => {
      setRemovingMember(null);
      queryClient.invalidateQueries({ queryKey: projectQueryKeys.all });
      queryClient.invalidateQueries({ queryKey: taskQueryKeys.all });
    },
  });

  const handleRemoveMember = (member: Member) => {
    setRemovingMember(member);
  };

  const confirmRemoveMember = () => {
    if (!project?.id || !removingMember?.id) return;

    mutate({
      projectId: project.id,
      memberUserId: removingMember.user_id,
    });
  };

  if (removingMember) {
    return (
      <Dialog open={!!removingMember} onOpenChange={() => setRemovingMember(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove member from project</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            Are you sure you want to remove {removingMember.user.name} from the project?
          </DialogDescription>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setRemovingMember(null)}>
              Cancel
            </Button>
            <Button onClick={confirmRemoveMember} disabled={isPending}>
              {isPending ? <LoadingSpinner size="1.5em" /> : 'Remove'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  const formatMemberCount = () => {
    if (project?.members.length === 0) {
      return 'No members';
    }

    if (project?.members.length === 1) {
      return '1 member';
    }

    return `${project?.members.length} members`;
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button
          type="button"
          className="flex w-fit items-center gap-2 rounded-md p-2 transition-colors hover:bg-slate-100 dark:hover:bg-slate-700"
          title="View project members"
        >
          <Users className="h-5 w-5 text-slate-500" />
        </button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Users className="h-5 w-5 text-slate-500" />
            Project members
          </DialogTitle>
          <DialogDescription>{formatMemberCount()}</DialogDescription>
        </DialogHeader>

        <section className="flex w-full min-w-0 flex-col gap-3 overflow-hidden border-t border-slate-200 pt-3 dark:border-slate-700">
          {project?.members.map((member) => (
            <article
              key={member.id}
              className="group hover:bg-primary/5 flex w-full min-w-0 items-center justify-between gap-2 rounded-md py-2 pr-4 pl-3"
            >
              <div className="flex min-w-0 flex-1 items-center gap-2">
                <div
                  className={cn(
                    'flex h-8 w-8 items-center justify-center rounded-full border-2 border-white bg-blue-600 text-xs font-medium text-white dark:border-slate-800',
                  )}
                  aria-hidden="true"
                >
                  {member.user.name.charAt(0).toUpperCase()}
                </div>
                <div className="flex flex-col gap-0.5 truncate">
                  <p className="truncate text-sm font-medium">{member.user.name}</p>
                  <p className="truncate text-xs text-slate-500 dark:text-slate-400">{member.user.email}</p>
                </div>
              </div>
              <div className="grid shrink-0 grid-cols-[1fr_24px] items-center gap-2">
                <div className="flex items-center gap-1 rounded-md border border-slate-900 px-2 py-0.5 text-sm whitespace-nowrap text-slate-500">
                  {memberRoleIcon[member.role]}
                  {memberRoleLabel[member.role]}
                </div>
                {isCreator && member.user_id !== userId && (
                  <Button
                    className="hover:bg-accent invisible bg-transparent px-2 text-slate-500 opacity-0 group-hover:visible group-hover:opacity-100 hover:text-red-500"
                    aria-label="Remove member from project"
                    onClick={() => handleRemoveMember(member)}
                  >
                    <Trash className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </article>
          ))}
        </section>
      </DialogContent>
    </Dialog>
  );
};

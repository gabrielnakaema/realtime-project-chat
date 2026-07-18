import { useMemo } from 'react';
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip';
import { cn } from '@/lib/utils';

interface Member {
  user_id: string;
  name: string;
}

interface MembersAvatarListProps {
  members: Member[];
  max?: number;
  onlineUserIds?: string[];
}

export const MembersAvatarList = ({ members = [], max = 4, onlineUserIds = [] }: MembersAvatarListProps) => {
  const sortedMembers = useMemo(() => {
    const membersCopy = members.map((member) => ({
      ...member,
      online: onlineUserIds.includes(member.user_id),
    }));

    const sorted = membersCopy.sort((a, b) => {
      return Number(b.online) - Number(a.online);
    });

    return sorted;
  }, [members, onlineUserIds]);

  const membersToShow = sortedMembers.slice(0, max);
  const remainingMembers = sortedMembers.slice(max);
  const remaining = sortedMembers.length - max;

  return (
    <div className="flex -space-x-2">
      {membersToShow.map((member) => (
        <Tooltip key={member.user_id}>
          <TooltipTrigger>
            <div
              className={cn(
                'border-card bg-primary text-primary-foreground flex h-8 w-8 items-center justify-center rounded-full border-2 text-xs font-medium',
                member.online && 'border-success',
              )}
            >
              {member.name.charAt(0).toUpperCase()}
            </div>
          </TooltipTrigger>
          <TooltipContent>{member.name}</TooltipContent>
        </Tooltip>
      ))}
      {remaining > 0 && (
        <Tooltip>
          <TooltipTrigger>
            <div className="border-card bg-muted flex h-8 w-8 items-center justify-center rounded-full border-2">
              <span className="text-muted-foreground text-xs font-medium">+{remaining}</span>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            {remainingMembers.map((member) => (
              <p key={member.user_id}>{member.name}</p>
            ))}
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );
};

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
                'w-8 h-8 bg-blue-600 rounded-full border-2 border-white dark:border-slate-800 flex items-center justify-center text-white text-xs font-medium',
                member.online && 'border-green-500 dark:border-green-500',
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
            <div className="w-8 h-8 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-white dark:border-slate-800 flex items-center justify-center">
              <span className="text-xs font-medium text-slate-600 dark:text-slate-400">+{remaining}</span>
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

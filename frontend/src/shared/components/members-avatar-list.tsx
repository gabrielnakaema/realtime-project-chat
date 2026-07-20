import { useMemo } from 'react';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip';
import { cn } from '@/lib/utils';

interface Member {
  user_id: string;
  name: string;
}

type AvatarVariant = 'default' | 'compactMuted';

interface MembersAvatarListProps {
  members: Member[];
  max?: number;
  onlineUserIds?: string[];
  variant?: AvatarVariant;
}

const variantStyles: Record<AvatarVariant, { list: string; avatar: string; badge: string; badgeText: string }> = {
  default: {
    list: '-space-x-2',
    avatar: 'bg-primary text-primary-foreground h-8 w-8 text-xs font-medium',
    badge: 'h-8 w-8',
    badgeText: 'text-xs',
  },
  compactMuted: {
    list: '-space-x-1.5',
    avatar: 'bg-muted text-muted-foreground size-7 text-[9px] font-semibold',
    badge: 'size-7',
    badgeText: 'text-[9px] font-semibold',
  },
};

const getInitials = (name: string, maxLetters: number) => {
  const words = name.trim().split(/\s+/).filter(Boolean);

  return words
    .slice(0, maxLetters)
    .map((word) => word.charAt(0).toUpperCase())
    .join('');
};

export const MembersAvatarList = ({
  members = [],
  max = 4,
  onlineUserIds = [],
  variant = 'default',
}: MembersAvatarListProps) => {
  const styles = variantStyles[variant];
  const initialLetters = variant === 'compactMuted' ? 2 : 1;

  const sortedMembers = useMemo(() => {
    const membersCopy = members.map((member) => ({
      ...member,
      online: onlineUserIds.includes(member.user_id),
    }));

    return membersCopy.sort((a, b) => Number(b.online) - Number(a.online));
  }, [members, onlineUserIds]);

  const membersToShow = sortedMembers.slice(0, max);
  const remainingMembers = sortedMembers.slice(max);
  const remaining = sortedMembers.length - max;

  return (
    <div className={cn('flex', styles.list)} aria-label={`${members.length} project members`}>
      {membersToShow.map((member) => (
        <Tooltip key={member.user_id}>
          <TooltipTrigger>
            <div
              className={cn(
                'border-card flex items-center justify-center rounded-full border-2',
                styles.avatar,
                member.online && 'border-success',
              )}
            >
              {getInitials(member.name, initialLetters)}
            </div>
          </TooltipTrigger>
          <TooltipContent>{member.name}</TooltipContent>
        </Tooltip>
      ))}
      {remaining > 0 && (
        <Tooltip>
          <TooltipTrigger>
            <div className={cn('border-card bg-muted flex items-center justify-center rounded-full border-2', styles.badge)}>
              <span className={cn('text-muted-foreground font-medium', styles.badgeText)}>+{remaining}</span>
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

import { cn } from '@/lib/utils';

interface UnreadCountBadgeProps {
  count: number;
  hasMoreUnread?: boolean;
  className?: string;
}

export const UnreadCountBadge = ({ count, hasMoreUnread = false, className }: UnreadCountBadgeProps) => {
  if (count <= 0) {
    return null;
  }

  const showOverflowStyle = hasMoreUnread || count > 99;

  return (
    <span
      className={cn(
        'rounded-full px-2 py-0.5 text-[10px] font-semibold',
        showOverflowStyle ? 'bg-destructive text-primary-foreground' : 'bg-primary text-primary-foreground',
        className,
      )}
    >
      {showOverflowStyle ? '+99' : String(count)}
    </span>
  );
};

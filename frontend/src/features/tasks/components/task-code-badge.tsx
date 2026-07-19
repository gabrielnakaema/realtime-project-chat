import { cn } from '@/lib/utils';

interface TaskCodeBadgeProps {
  code?: string | null;
  className?: string;
}

export const TaskCodeBadge = ({ code, className }: TaskCodeBadgeProps) => {
  const normalizedCode = code?.trim();

  if (!normalizedCode) {
    return null;
  }

  return (
    <span
      className={cn(
        'text-muted-foreground inline-flex max-w-full min-w-0 items-center overflow-hidden rounded-md font-mono text-[10px] leading-none font-semibold tracking-[0.02em]',
        className,
      )}
      title={normalizedCode}
    >
      <span className="truncate">{normalizedCode}</span>
    </span>
  );
};

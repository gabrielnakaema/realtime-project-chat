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
        'inline-flex max-w-full min-w-0 items-center overflow-hidden rounded-md border border-slate-300 bg-slate-100 px-2 py-1 font-mono text-[10px] leading-none font-semibold tracking-[0.08em] text-slate-600 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300',
        className,
      )}
      title={normalizedCode}
    >
      <span className="truncate">{normalizedCode}</span>
    </span>
  );
};

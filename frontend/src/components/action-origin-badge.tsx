import { Bot } from 'lucide-react';
import type { ActionOrigin } from '@/types/action-origin';
import { cn } from '@/lib/utils';
import { isMCPAgentAction } from '@/types/action-origin';

interface ActionOriginBadgeProps {
  origin?: ActionOrigin | null;
  className?: string;
}

export const ActionOriginBadge = ({ origin, className }: ActionOriginBadgeProps) => {
  if (!isMCPAgentAction(origin)) {
    return null;
  }

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:border-emerald-900/80 dark:bg-emerald-950/40 dark:text-emerald-300',
        className,
      )}
      title="Action performed by AI agent via MCP"
      aria-label="Action performed by AI agent via MCP"
    >
      <Bot className="h-3 w-3" aria-hidden="true" />
      AI Agent
    </span>
  );
};

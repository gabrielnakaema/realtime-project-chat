import { Bot } from 'lucide-react';
import type { ActionOrigin } from '@/features/tasks/types/action-origin';
import { cn } from '@/lib/utils';
import { isMCPAgentAction } from '@/features/tasks/types/action-origin';

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
        'border-success/30 bg-success/10 text-success inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] font-medium',
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

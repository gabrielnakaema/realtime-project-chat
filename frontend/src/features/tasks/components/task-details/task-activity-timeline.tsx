import { Activity, ChevronDown } from 'lucide-react';
import { useState } from 'react';
import { TaskDetailsUpdate } from './task-details-update';
import type { TaskUpdate } from '@/features/tasks/types/task';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/shared/components/ui/collapsible';
import { useAuth } from '@/features/auth/hooks/use-auth';
import { cn } from '@/lib/utils';

interface TaskActivityTimelineProps {
  updates: TaskUpdate[];
}

export const TaskActivityTimeline = ({ updates = [] }: TaskActivityTimelineProps) => {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);

  return (
    <Collapsible open={open} onOpenChange={setOpen} className="border-border rounded-2xl border">
      <CollapsibleTrigger className="hover:bg-muted flex w-full items-center justify-between gap-3 px-5 py-4 text-left transition-colors">
        <div className="flex items-center gap-3">
          <div className="bg-muted rounded-full p-2">
            <Activity className="text-muted-foreground h-4 w-4" />
          </div>
          <div>
            <p className="text-foreground text-sm font-semibold">Activity timeline</p>
            <p className="text-muted-foreground text-xs">
              {updates.length > 0 ? `${updates.length} updates` : 'No activity yet'}
            </p>
          </div>
        </div>
        <ChevronDown className={cn('text-muted-foreground h-4 w-4 transition-transform', open && 'rotate-180')} />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="border-border border-t px-5 py-4">
          {updates.length > 0 && (
            <div className="flex flex-col">
              {updates.map((update, index) => (
                <TaskDetailsUpdate
                  key={update.id}
                  update={update}
                  isLast={index === updates.length - 1}
                  currentUserId={user?.id}
                />
              ))}
            </div>
          )}
          {updates.length === 0 && (
            <p className="text-muted-foreground text-sm">Activity will appear here as the task changes.</p>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
};

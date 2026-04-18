import { Activity, ChevronDown } from 'lucide-react';
import { useState } from 'react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '../ui/collapsible';
import { TaskDetailsUpdate } from './task-details-update';
import type { TaskUpdate } from '@/types/task';
import { cn } from '@/lib/utils';

interface TaskActivityTimelineProps {
  updates: TaskUpdate[];
}

export const TaskActivityTimeline = ({ updates }: TaskActivityTimelineProps) => {
  const [open, setOpen] = useState(false);

  return (
    <Collapsible
      open={open}
      onOpenChange={setOpen}
      className="rounded-2xl border border-slate-200 dark:border-slate-700"
    >
      <CollapsibleTrigger className="flex w-full items-center justify-between gap-3 px-5 py-4 text-left transition-colors hover:bg-slate-50 dark:hover:bg-slate-900/50">
        <div className="flex items-center gap-3">
          <div className="rounded-full bg-slate-100 p-2 dark:bg-slate-800">
            <Activity className="h-4 w-4 text-slate-600 dark:text-slate-300" />
          </div>
          <div>
            <p className="text-sm font-semibold text-slate-800 dark:text-slate-100">Activity timeline</p>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {updates.length > 0 ? `${updates.length} updates` : 'No activity yet'}
            </p>
          </div>
        </div>
        <ChevronDown
          className={cn('h-4 w-4 text-slate-500 transition-transform dark:text-slate-400', open && 'rotate-180')}
        />
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="border-t border-slate-200 px-5 py-4 dark:border-slate-700">
          {updates.length > 0 && (
            <div className="flex flex-col">
              {updates.map((update, index) => (
                <TaskDetailsUpdate key={update.id} update={update} isLast={index === updates.length - 1} />
              ))}
            </div>
          )}
          {updates.length === 0 && (
            <p className="text-sm text-slate-500 dark:text-slate-400">Activity will appear here as the task changes.</p>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
};

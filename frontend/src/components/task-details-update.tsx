import { useState } from 'react';
import { ChevronRight } from 'lucide-react';
import { Avatar } from './avatar';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from './ui/collapsible';
import type { TaskUpdate } from '@/types/task';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';
import { cn } from '@/lib/utils';

const taskFieldsLabels: Record<string, string> = {
  title: 'Title',
  description: 'Description',
  status: 'Status',
  priority: 'Priority',
  responsible_id: 'Responsible',
  due_date: 'Due date',
  done_at: 'Done at',
  created_at: 'Created at',
  updated_at: 'Updated at',
  tags: 'Tags',
  author_id: 'Author',
};

export const TaskDetailsUpdate = ({ update }: { update: TaskUpdate }) => {
  const [open, setOpen] = useState(false);

  const dateClassNames = 'text-xs text-slate-500 dark:text-slate-400';
  const changeTextClassNames = 'text-xs text-slate-500 dark:text-slate-400';

  const hasChanges = !!update.changes?.length;

  return (
    <Collapsible className="w-full" open={open} onOpenChange={setOpen} disabled={!hasChanges}>
      <div key={update.id} className="flex w-full items-center gap-3">
        {update.user.name && <Avatar name={update.user.name} size="sm" />}

        <CollapsibleTrigger asChild>
          <button
            type="button"
            className={cn(
              'flex flex-1 items-center justify-between gap-8 rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-900',
            )}
          >
            <div className="flex flex-col items-start gap-1">
              <UpdateText update={update} />
              <p className={dateClassNames}>{formatRelativeActivityDateString(update.created_at)}</p>
            </div>
            {update.update_type === 'updated' && !!update.changes?.length && (
              <div
                className={cn(
                  'text-sm text-slate-500 transition-transform duration-200 dark:text-slate-400',
                  open && 'rotate-90',
                )}
              >
                <ChevronRight />
              </div>
            )}
          </button>
        </CollapsibleTrigger>
      </div>
      <CollapsibleContent>
        <div className="flex flex-col gap-0.5 pt-2 pl-12">
          <p className={changeTextClassNames}>
            Updated{' '}
            <span>{update.changes?.map((change) => taskFieldsLabels[change.field] || change.field).join(', ')}</span>
          </p>
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
};

const GeneralText = ({ children }: { children: React.ReactNode }) => {
  return <p className="text-sm text-slate-500 dark:text-slate-400">{children}</p>;
};

const Actor = ({ children }: { children: React.ReactNode }) => {
  return <span className="font-semibold">{children}</span>;
};

const Subject = ({ children }: { children: React.ReactNode }) => {
  return <span className="font-semibold">{children}</span>;
};

const UpdateText = ({ update }: { update: TaskUpdate }) => {
  if (update.update_type === 'created') {
    return (
      <GeneralText>
        <Actor>{update.user.name}</Actor> created the task
      </GeneralText>
    );
  }

  if (update.update_type === 'assigned') {
    return (
      <GeneralText>
        <Actor>{update.user.name}</Actor> assigned the task to{' '}
        <Subject>{update.changes?.[0]?.subject?.name || 'someone'}</Subject>
      </GeneralText>
    );
  }

  if (update.update_type === 'unassigned') {
    return (
      <GeneralText>
        <Actor>{update.user.name}</Actor> removed the current assignee
      </GeneralText>
    );
  }

  if (update.update_type === 'status') {
    return (
      <GeneralText>
        <Actor>{update.user.name}</Actor> updated status to <Subject>{update.changes?.[0]?.new_value}</Subject>
      </GeneralText>
    );
  }

  if (update.update_type === 'done') {
    return (
      <GeneralText>
        <Actor>{update.user.name}</Actor> marked as done
      </GeneralText>
    );
  }

  return (
    <GeneralText>
      <Actor>{update.user.name}</Actor> {update.update_type} the task
    </GeneralText>
  );
};

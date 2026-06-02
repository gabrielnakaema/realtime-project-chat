import { ArrowRight } from 'lucide-react';
import { Avatar } from '../avatar';
import { TaskPriorityBadge } from '../task-priority-badge';
import { TaskStatusBadge } from '../task-status-badge';
import type { TaskChange, TaskPriority, TaskUpdate } from '@/types/task';
import { formatTaskDueDate } from '@/utils/date';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';

const fieldLabels: Record<string, string> = {
  title: 'Title',
  description: 'Description',
  column: 'Column',
  priority: 'Priority',
  responsible_id: 'Responsible',
  due_date: 'Due date',
  done_at: 'Done at',
};

export const TaskDetailsUpdate = ({ update, isLast }: { update: TaskUpdate; isLast: boolean }) => {
  return (
    <div className="relative flex gap-3">
      {!isLast && <div className="absolute top-6 bottom-0 left-[11px] w-px bg-slate-200 dark:bg-slate-700" />}
      <div className="relative z-10 mt-0.5 shrink-0">
        <Avatar name={update.user.name} size="sm" />
      </div>
      <div className="flex min-w-0 flex-1 flex-col gap-0.5 pb-5">
        <UpdateSummary update={update} />
        <span className="text-xs text-slate-400 dark:text-slate-500">
          {formatRelativeActivityDateString(update.created_at)}
        </span>
        <UpdateChanges update={update} />
      </div>
    </div>
  );
};

const Actor = ({ name }: { name: string }) => (
  <span className="font-medium text-slate-900 dark:text-slate-100">{name}</span>
);

const UpdateSummary = ({ update }: { update: TaskUpdate }) => {
  const changes = update.changes ?? [];
  const responsibleChange = changes.find((change) => change.field === 'responsible_id');

  if (update.update_type === 'created') {
    return (
      <p className="text-sm text-slate-600 dark:text-slate-300">
        <Actor name={update.user.name} /> created the task
      </p>
    );
  }

  if (update.update_type === 'assigned') {
    const subject = responsibleChange?.subject?.name ?? getDisplayValue(responsibleChange, 'new');
    return (
      <p className="text-sm text-slate-600 dark:text-slate-300">
        <Actor name={update.user.name} /> assigned to{' '}
        <span className="font-medium text-slate-900 dark:text-slate-100">{subject}</span>
      </p>
    );
  }

  if (update.update_type === 'unassigned') {
    return (
      <p className="text-sm text-slate-600 dark:text-slate-300">
        <Actor name={update.user.name} /> removed the assignee
      </p>
    );
  }

  if (update.update_type === 'done') {
    return (
      <p className="text-sm text-slate-600 dark:text-slate-300">
        <Actor name={update.user.name} /> marked as done
      </p>
    );
  }

  const isColumnOnly =
    changes.length === 1 &&
    changes[0].field === 'column' &&
    (update.update_type === 'column' || update.update_type === 'updated');

  if (isColumnOnly) {
    return (
      <p className="flex flex-wrap items-center gap-1.5 text-sm text-slate-600 dark:text-slate-300">
        <Actor name={update.user.name} /> moved to{' '}
        <TaskStatusBadge status={changes[0].new_value} label={changes[0].new_display_value ?? changes[0].new_value} />
      </p>
    );
  }

  const isPriorityOnly = changes.length === 1 && changes[0].field === 'priority' && update.update_type === 'updated';

  if (isPriorityOnly) {
    return (
      <p className="flex flex-wrap items-center gap-1.5 text-sm text-slate-600 dark:text-slate-300">
        <Actor name={update.user.name} /> set priority to{' '}
        <TaskPriorityBadge priority={changes[0].new_value as TaskPriority} />
      </p>
    );
  }

  const fieldNames = changes.map((c) => fieldLabels[c.field] ?? c.field).join(', ');

  return (
    <p className="text-sm text-slate-600 dark:text-slate-300">
      <Actor name={update.user.name} /> updated <span className="text-slate-700 dark:text-slate-200">{fieldNames}</span>
    </p>
  );
};

const UpdateChanges = ({ update }: { update: TaskUpdate }) => {
  const changes = update.changes ?? [];

  if (changes.length === 0) return null;

  const skipTypes = new Set(['created', 'assigned', 'unassigned', 'done']);
  if (skipTypes.has(update.update_type)) return null;

  if (changes.length === 1 && (changes[0].field === 'column' || changes[0].field === 'priority')) return null;

  return (
    <div className="mt-1.5 flex flex-col gap-1.5">
      {changes.map((change) => (
        <ChangeRow key={change.id} change={change} />
      ))}
    </div>
  );
};

const ChangeRow = ({ change }: { change: TaskChange }) => {
  if (change.field === 'column') {
    return (
      <div className="flex items-center gap-1.5">
        <span className="w-20 shrink-0 text-xs text-slate-400 dark:text-slate-500">{fieldLabels.column}</span>
        <TaskStatusBadge status={change.old_value} label={change.old_display_value ?? change.old_value} />
        <ArrowRight className="h-3 w-3 shrink-0 text-slate-400" />
        <TaskStatusBadge status={change.new_value} label={change.new_display_value ?? change.new_value} />
      </div>
    );
  }

  if (change.field === 'priority') {
    return (
      <div className="flex items-center gap-1.5">
        <span className="w-20 shrink-0 text-xs text-slate-400 dark:text-slate-500">{fieldLabels.priority}</span>
        <TaskPriorityBadge priority={change.old_value as TaskPriority} />
        <ArrowRight className="h-3 w-3 shrink-0 text-slate-400" />
        <TaskPriorityBadge priority={change.new_value as TaskPriority} />
      </div>
    );
  }

  if (change.field === 'description') {
    return (
      <p className="text-xs text-slate-400 dark:text-slate-500">
        <span className="w-20 shrink-0">{fieldLabels.description}</span> was updated
      </p>
    );
  }

  if (change.field === 'due_date') {
    return (
      <div className="flex items-center gap-1.5">
        <span className="w-20 shrink-0 text-xs text-slate-400 dark:text-slate-500">{fieldLabels.due_date}</span>
        <span className="text-xs text-slate-400 dark:text-slate-500">{formatTaskDueDate(change.old_value)}</span>
        <ArrowRight className="h-3 w-3 shrink-0 text-slate-400" />
        <span className="text-xs text-slate-700 dark:text-slate-200">{formatTaskDueDate(change.new_value)}</span>
      </div>
    );
  }

  const label = fieldLabels[change.field] ?? change.field;
  const oldValue = getDisplayValue(change, 'old');
  const newValue = getDisplayValue(change, 'new');
  const isLong = oldValue.length > 40 || newValue.length > 40;

  if (isLong) {
    return (
      <div className="flex flex-col gap-0.5">
        <span className="text-xs text-slate-400 dark:text-slate-500">{label}</span>
        <span className="text-xs text-slate-400 line-through">{oldValue || '—'}</span>
        <span className="text-xs text-slate-700 dark:text-slate-200">{newValue || '—'}</span>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400">
      <span className="w-20 shrink-0">{label}</span>
      <span className="line-through">{oldValue || '—'}</span>
      <ArrowRight className="h-3 w-3 shrink-0 text-slate-400" />
      <span className="text-slate-700 dark:text-slate-200">{newValue || '—'}</span>
    </div>
  );
};

const getDisplayValue = (change: TaskChange | undefined, direction: 'old' | 'new') => {
  if (!change) {
    return '';
  }

  if (direction === 'old') {
    return change.old_display_value ?? change.old_value;
  }

  return change.new_display_value ?? change.new_value;
};

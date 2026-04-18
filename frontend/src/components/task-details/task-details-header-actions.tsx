import { Pencil, Trash2 } from 'lucide-react';
import { useState } from 'react';
import type { TaskStatus } from '@/types/task';

interface TaskDetailsHeaderActionsProps {
  status: TaskStatus;
  isArchiving: boolean;
  onEdit: () => void;
  onArchive: () => void;
}

export const TaskDetailsHeaderActions = ({ status, isArchiving, onEdit, onArchive }: TaskDetailsHeaderActionsProps) => {
  const [isConfirmingArchive, setIsConfirmingArchive] = useState(false);

  if (isConfirmingArchive) {
    return (
      <>
        <span className="text-xs text-slate-500 dark:text-slate-400">Archive task?</span>
        <button
          type="button"
          className="rounded-md px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950"
          onClick={onArchive}
          disabled={isArchiving}
        >
          {isArchiving ? 'Archiving...' : 'Confirm'}
        </button>
        <button
          type="button"
          className="rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
          onClick={() => setIsConfirmingArchive(false)}
        >
          Cancel
        </button>
      </>
    );
  }

  return (
    <>
      <button
        type="button"
        className="w-fit rounded-md p-2 font-medium transition-colors hover:bg-slate-100 dark:hover:bg-slate-800"
        onClick={onEdit}
      >
        <Pencil className="h-4 w-4 text-slate-500 dark:text-slate-400" />
      </button>
      {status !== 'archived' && (
        <button
          type="button"
          className="w-fit rounded-md p-2 font-medium transition-colors hover:bg-red-50 dark:hover:bg-red-950"
          onClick={() => setIsConfirmingArchive(true)}
        >
          <Trash2 className="h-4 w-4 text-slate-500 dark:text-slate-400" />
        </button>
      )}
    </>
  );
};

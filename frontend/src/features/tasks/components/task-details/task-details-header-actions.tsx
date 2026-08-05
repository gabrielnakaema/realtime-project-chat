import { Pencil, Trash2 } from 'lucide-react';
import { useState } from 'react';

interface TaskDetailsHeaderActionsProps {
  archivedAt: string | null;
  isArchiving: boolean;
  onEdit: () => void;
  onArchive: () => void;
}

export const TaskDetailsHeaderActions = ({
  archivedAt,
  isArchiving,
  onEdit,
  onArchive,
}: TaskDetailsHeaderActionsProps) => {
  const [isConfirmingArchive, setIsConfirmingArchive] = useState(false);

  if (isConfirmingArchive) {
    return (
      <>
        <span className="text-muted-foreground text-xs">Archive task?</span>
        <button
          type="button"
          className="text-destructive hover:bg-destructive/10 rounded-md px-2 py-1 text-xs font-medium transition-colors"
          onClick={onArchive}
          disabled={isArchiving}
        >
          {isArchiving ? 'Archiving...' : 'Confirm'}
        </button>
        <button
          type="button"
          className="text-muted-foreground hover:bg-muted rounded-md px-2 py-1 text-xs font-medium transition-colors"
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
        aria-label="Edit task"
        className="hover:bg-muted w-fit rounded-md p-2 font-medium transition-colors"
        onClick={onEdit}
      >
        <Pencil className="text-muted-foreground h-4 w-4" />
      </button>
      {archivedAt === null && (
        <button
          type="button"
          aria-label="Archive task"
          className="hover:bg-destructive/10 w-fit rounded-md p-2 font-medium transition-colors"
          onClick={() => setIsConfirmingArchive(true)}
        >
          <Trash2 className="text-muted-foreground h-4 w-4" />
        </button>
      )}
    </>
  );
};

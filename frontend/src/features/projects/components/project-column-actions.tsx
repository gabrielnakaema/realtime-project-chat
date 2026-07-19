import { MoreHorizontal, PencilLine, Plus } from 'lucide-react';
import { useState } from 'react';
import { EditColumnDialog } from './edit-column-dialog';
import type { Column } from '@/features/tasks/types/board';
import { CreateTask } from '@/features/tasks/components/task-form/create-task';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/shared/components/ui/dropdown-menu';

interface ProjectColumnActionsProps {
  column: Column;
  canEditColumn: boolean;
  surfaceAccentColor: string;
}

export const ProjectColumnActions = ({ column, canEditColumn, surfaceAccentColor }: ProjectColumnActionsProps) => {
  const [isCreateTaskOpen, setIsCreateTaskOpen] = useState(false);
  const [isEditColumnOpen, setIsEditColumnOpen] = useState(false);

  return (
    <>
      <CreateTask
        projectId={column.project_id}
        initialProjectColumnId={column.columnId}
        open={isCreateTaskOpen}
        onOpenChange={setIsCreateTaskOpen}
        trigger={null}
      />
      <EditColumnDialog column={column} open={isEditColumnOpen} onOpenChange={setIsEditColumnOpen} />

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button
            type="button"
            className="text-muted-foreground hover:text-foreground rounded p-1 transition-colors"
            style={{ color: surfaceAccentColor }}
            aria-label={`Open actions for ${column.title}`}
          >
            <MoreHorizontal className="h-4 w-4" />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            onSelect={() => {
              setIsCreateTaskOpen(true);
            }}
          >
            <Plus className="h-4 w-4" />
            Add task
          </DropdownMenuItem>
          {canEditColumn && (
            <DropdownMenuItem
              onSelect={() => {
                setIsEditColumnOpen(true);
              }}
            >
              <PencilLine className="h-4 w-4" />
              Edit column
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
};

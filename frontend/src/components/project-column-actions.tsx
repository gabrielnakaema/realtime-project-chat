import { MoreHorizontal, PencilLine, Plus } from 'lucide-react';
import { useState } from 'react';
import { CreateTask } from './task-form/create-task';
import { EditColumnDialog } from './edit-column-dialog';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu';
import type { Column } from '@/types/board';

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
            className="rounded p-1 text-slate-500 transition-colors hover:text-slate-700 dark:hover:bg-slate-800"
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

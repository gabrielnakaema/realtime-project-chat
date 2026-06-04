import { Archive } from 'lucide-react';
import { useState } from 'react';
import { TaskDetails } from '../task-details';
import { EditTask } from '../task-form/edit-task';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../ui/dialog';
import { ScrollArea } from '../ui/scroll-area';
import { ArchivedTasksList } from './archived-tasks-list';
import type { Project } from '@/types/project';

interface ArchivedTasksModalProps {
  project: Project;
}

export const ArchivedTasksModal = ({ project }: ArchivedTasksModalProps) => {
  const [open, setOpen] = useState(false);
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [editingTaskId, setEditingTaskId] = useState<string | null>(null);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:text-slate-500 dark:hover:bg-slate-800 dark:hover:text-slate-300"
      >
        <Archive className="h-3.5 w-3.5" />
        Archived
      </button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="md:max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-base">
              <Archive className="h-4 w-4 text-slate-400" />
              Archived tasks
            </DialogTitle>
          </DialogHeader>

          <ScrollArea className="max-h-[60vh]">
            <ArchivedTasksList project={project} open={open} onSelectTask={setSelectedTaskId} />
          </ScrollArea>
        </DialogContent>
      </Dialog>

      {selectedTaskId && (
        <TaskDetails
          taskId={selectedTaskId}
          open={!!selectedTaskId}
          onOpenChange={(value) => {
            if (!value) {
              setSelectedTaskId(null);
            }
          }}
          onEdit={() => {
            if (!selectedTaskId) {
              return;
            }

            setEditingTaskId(selectedTaskId);
            setSelectedTaskId(null);
          }}
        />
      )}
      {editingTaskId && (
        <EditTask
          taskId={editingTaskId}
          open={!!editingTaskId}
          onOpenChange={(value) => {
            if (!value) {
              setEditingTaskId(null);
            }
          }}
        />
      )}
    </>
  );
};

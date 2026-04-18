import { Dialog, DialogContent } from '../ui/dialog';
import type { ReactNode } from 'react';

interface TaskDetailsStateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
}

export const TaskDetailsStateDialog = ({ open, onOpenChange, children }: TaskDetailsStateDialogProps) => {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <div className="flex min-h-50 flex-col items-center justify-center">{children}</div>
      </DialogContent>
    </Dialog>
  );
};

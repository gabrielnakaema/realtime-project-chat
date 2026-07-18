import { MessageSquare } from 'lucide-react';

export const EmptyTaskCommentsState = () => {
  return (
    <div className="border-border flex flex-col items-center justify-center rounded-2xl border border-dashed px-6 py-12 text-center">
      <MessageSquare className="text-muted-foreground mb-3 h-8 w-8" />
      <p className="text-foreground text-sm font-medium">No comments yet</p>
      <p className="text-muted-foreground mt-1 max-w-sm text-sm">
        Use comments to capture decisions, ask for feedback, and keep discussion attached to the task.
      </p>
    </div>
  );
};

import { MessageSquare } from 'lucide-react';

export const EmptyTaskCommentsState = () => {
  return (
    <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-300 px-6 py-12 text-center dark:border-slate-700">
      <MessageSquare className="mb-3 h-8 w-8 text-slate-400 dark:text-slate-500" />
      <p className="text-sm font-medium text-slate-700 dark:text-slate-200">No comments yet</p>
      <p className="mt-1 max-w-sm text-sm text-slate-500 dark:text-slate-400">
        Use comments to capture decisions, ask for feedback, and keep discussion attached to the task.
      </p>
    </div>
  );
};

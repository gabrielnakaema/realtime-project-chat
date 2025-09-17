import { forwardRef } from 'react';

interface TextareaProps extends React.ComponentPropsWithoutRef<'textarea'> {
  label: string;
  error?: string;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>((props, ref) => {
  return (
    <div className="space-y-2">
      <label htmlFor={props.id} className="block text-sm font-medium text-slate-700 dark:text-slate-300">
        {props.label}
      </label>
      <textarea
        id={props.id}
        className="w-full min-h-32 px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-500 dark:placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        {...props}
        ref={ref}
      />
      {props.error && <p className="text-sm text-red-500">{props.error}</p>}
    </div>
  );
});

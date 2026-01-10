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
        className="min-h-32 w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-slate-900 placeholder-slate-500 focus:border-transparent focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 dark:placeholder-slate-400"
        {...props}
        ref={ref}
      />
      {props.error && <p className="text-sm text-red-500">{props.error}</p>}
    </div>
  );
});

import { forwardRef } from 'react';

interface TextareaProps extends React.ComponentPropsWithoutRef<'textarea'> {
  label: string;
  error?: string;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>((props, ref) => {
  return (
    <div className="space-y-2">
      <label htmlFor={props.id} className="text-foreground block text-sm font-medium">
        {props.label}
      </label>
      <textarea
        id={props.id}
        className="border-border bg-card text-foreground placeholder-muted-foreground focus:ring-ring min-h-32 w-full rounded-md border px-3 py-2 focus:border-transparent focus:ring-2 focus:outline-none"
        {...props}
        ref={ref}
      />
      {props.error && <p className="text-destructive text-sm">{props.error}</p>}
    </div>
  );
});

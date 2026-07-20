import { forwardRef } from 'react';

interface InputProps extends React.ComponentPropsWithoutRef<'input'> {
  label: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>((props, ref) => {
  return (
    <div className="w-full space-y-2">
      {props.label && (
        <label htmlFor={props.id} className="text-foreground block text-sm font-medium">
          {props.label}
        </label>
      )}
      <input
        id={props.id}
        type={props.type}
        placeholder={props.placeholder}
        className="border-border bg-card text-foreground placeholder-muted-foreground focus:ring-ring w-full rounded-md border px-3 py-2 focus:border-transparent focus:ring-2 focus:outline-none"
        {...props}
        ref={ref}
      />
      {props.error && <p className="text-destructive text-sm">{props.error}</p>}
    </div>
  );
});

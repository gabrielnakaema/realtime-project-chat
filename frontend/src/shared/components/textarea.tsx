import { forwardRef } from 'react';
import { cn } from '@/lib/utils';

interface TextareaProps extends React.ComponentPropsWithoutRef<'textarea'> {
  label?: string;
  error?: string;
  classNames?: {
    label?: string;
    container?: string;
    input?: string;
    error?: string;
  };
}

export const Input = forwardRef<HTMLTextAreaElement, TextareaProps>((props, ref) => {
  const { label, error, classNames, id, ...rest } = props;
  const errorId = error && id ? `${id}-error` : undefined;

  return (
    <div className={cn('flex w-full flex-col gap-1.5', classNames?.container)}>
      {label && (
        <label htmlFor={id} className={cn('text-[11px] font-semibold tracking-wider', classNames?.label)}>
          {label}
        </label>
      )}
      <textarea
        {...rest}
        id={id}
        ref={ref}
        className={cn(
          'border-border bg-card text-foreground placeholder:text-muted-foreground h-[39px] w-full rounded-md border px-3 text-[13px]',
          classNames?.input,
        )}
        aria-invalid={!!error || undefined}
        aria-describedby={errorId}
      />
      {error && (
        <span role="alert" id={errorId} className={cn('text-destructive text-xs', classNames?.error)}>
          {error}
        </span>
      )}
    </div>
  );
});

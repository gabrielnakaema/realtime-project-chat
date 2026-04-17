import { forwardRef } from 'react';
import type { ComponentPropsWithoutRef } from 'react';
import { cn } from '@/lib/utils';

interface ExpandingTextareaProps extends ComponentPropsWithoutRef<'textarea'> {
  wrapperClassName?: string;
}

export const ExpandingTextarea = forwardRef<HTMLTextAreaElement, ExpandingTextareaProps>(
  ({ className, wrapperClassName, value = '', ...props }, ref) => {
    return (
      <div className={cn('relative', wrapperClassName)}>
        <div aria-hidden className={cn('invisible break-words whitespace-pre-wrap', className)}>
          {String(value) + '\n'}
        </div>
        <textarea
          ref={ref}
          rows={1}
          value={value}
          className={cn('absolute inset-0 h-full w-full resize-none overflow-hidden', className)}
          {...props}
        />
      </div>
    );
  },
);

ExpandingTextarea.displayName = 'ExpandingTextarea';

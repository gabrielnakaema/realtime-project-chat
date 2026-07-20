import { forwardRef } from 'react';
import type { ComponentPropsWithoutRef } from 'react';
import { cn } from '@/lib/utils';

interface ExpandingTextareaProps extends ComponentPropsWithoutRef<'textarea'> {
  wrapperClassName?: string;
}

export const ExpandingTextarea = forwardRef<HTMLTextAreaElement, ExpandingTextareaProps>(
  ({ className, wrapperClassName, value = '', ...props }, ref) => {
    return (
      <div className={cn('relative min-w-0', wrapperClassName)}>
        <div
          aria-hidden
          className={cn(
            'invisible min-w-0 [overflow-wrap:anywhere] [word-break:break-word] whitespace-pre-wrap',
            className,
          )}
        >
          {String(value) + '\n'}
        </div>
        <textarea
          ref={ref}
          rows={1}
          value={value}
          wrap="soft"
          className={cn(
            'absolute inset-0 h-full w-full min-w-0 resize-none overflow-hidden overflow-x-hidden [overflow-wrap:anywhere] [word-break:break-word] whitespace-pre-wrap',
            className,
          )}
          {...props}
        />
      </div>
    );
  },
);

ExpandingTextarea.displayName = 'ExpandingTextarea';

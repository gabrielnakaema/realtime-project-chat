import { forwardRef } from 'react';
import { cn } from '@/lib/utils';

interface ButtonProps extends React.ComponentPropsWithoutRef<'button'> {
  variant?: 'primary' | 'secondary';
}

const variantClassnames = {
  primary: 'bg-primary text-primary-foreground hover:bg-primary/90 disabled:bg-primary/50',
  secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80 disabled:bg-secondary/50',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', className, ...props }, ref) => {
    return (
      <button
        {...props}
        ref={ref}
        className={cn(
          'flex w-fit items-center justify-center gap-2 rounded-md px-4 py-2 font-medium transition-colors',
          variantClassnames[variant],
          className,
        )}
      >
        {props.children}
      </button>
    );
  },
);

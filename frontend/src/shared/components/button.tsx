import { forwardRef } from 'react';
import { cva } from 'class-variance-authority';
import type { VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const buttonVariants = cva(
  'inline-flex w-fit shrink-0 cursor-pointer items-center justify-center gap-2 rounded-md font-medium whitespace-nowrap transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90 disabled:bg-primary/50',
        outline: 'border border-border bg-background text-foreground hover:bg-muted',
        ghost: 'text-foreground hover:bg-muted',
        destructive: 'bg-destructive text-primary-foreground hover:bg-destructive/90 disabled:bg-destructive/50',
        destructiveOutline: 'border border-destructive/50 bg-background text-destructive hover:bg-destructive/10',
        link: 'text-foreground underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-9 px-4 py-2 text-sm',
        sm: 'h-8 px-3 text-xs',
        icon: 'size-9 p-0',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
);

type ButtonProps = React.ComponentPropsWithoutRef<'button'> & VariantProps<typeof buttonVariants>;

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'default', ...props }, ref) => {
    return (
      <button data-slot="button" className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />
    );
  },
);

Button.displayName = 'Button';

export { Button, buttonVariants };

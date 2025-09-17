import { cn } from '@/lib/utils';
import { forwardRef } from 'react';

interface ButtonProps extends React.ComponentPropsWithoutRef<'button'> {
  variant?: 'primary' | 'secondary';
}

const variantClassnames = {
  primary: 'text-white bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400',
  secondary:
    'text-white bg-gray-600 hover:bg-gray-700 disabled:bg-gray-400 dark:bg-gray-700 dark:hover:bg-gray-600 dark:disabled:bg-gray-500',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', className, ...props }, ref) => {
    return (
      <button
        {...props}
        ref={ref}
        className={cn(
          'w-fit px-4 py-2 rounded-md font-medium transition-colors flex items-center justify-center gap-2',
          variantClassnames[variant],
          className,
        )}
      >
        {props.children}
      </button>
    );
  },
);

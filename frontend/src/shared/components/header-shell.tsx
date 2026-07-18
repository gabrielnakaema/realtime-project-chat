import type { ComponentProps } from 'react';
import { cn } from '@/lib/utils';

type HeaderShellProps = ComponentProps<'header'> & {
  width?: 'contained' | 'fluid';
};

export const HeaderShell = ({ width = 'fluid', className, children, ...props }: HeaderShellProps) => {
  return (
    <header className={cn('border-border bg-card h-16 shrink-0 border-b', className)} {...props}>
      <div
        className={cn(
          'mx-auto flex h-full w-full min-w-0 items-center gap-3 px-4 sm:px-6',
          width === 'contained' && 'max-w-6xl',
        )}
      >
        {children}
      </div>
    </header>
  );
};

export const HeaderStart = ({ className, ...props }: ComponentProps<'div'>) => {
  return <div className={cn('flex min-w-0 items-center gap-2', className)} {...props} />;
};

export const HeaderCenter = ({ className, ...props }: ComponentProps<'div'>) => {
  return <div className={cn('flex min-w-0 flex-1 items-center', className)} {...props} />;
};

export const HeaderActions = ({ className, ...props }: ComponentProps<'div'>) => {
  return <div className={cn('ml-auto flex shrink-0 items-center gap-2', className)} {...props} />;
};

type HeaderLogoProps = Omit<ComponentProps<'img'>, 'alt' | 'src'>;

export const HeaderLogo = ({ className, ...props }: HeaderLogoProps) => {
  return <img src="/logo.png" alt="" className={cn('size-8 shrink-0 object-contain', className)} {...props} />;
};

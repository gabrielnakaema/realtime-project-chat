import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip';
import { cn } from '@/lib/utils';

interface AvatarProps {
  name: string;
  size?: 'sm' | 'md' | 'lg';
  containerClassName?: string;
}

export const Avatar = ({ name, size = 'sm', containerClassName }: AvatarProps) => {
  const sizeClassNames = {
    sm: 'h-6 w-6 min-w-6 min-h-6',
    md: 'h-8 w-8 min-w-8 min-h-8',
    lg: 'h-10 w-10 min-w-10 min-h-10',
  };

  return (
    <Tooltip>
      <TooltipTrigger>
        <div
          className={cn(
            'bg-primary text-primary-foreground flex items-center justify-center rounded-full text-xs font-medium',
            sizeClassNames[size],
            containerClassName,
          )}
        >
          {name.charAt(0).toUpperCase()}
        </div>
      </TooltipTrigger>
      <TooltipContent>{name}</TooltipContent>
    </Tooltip>
  );
};

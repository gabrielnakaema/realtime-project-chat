import { LoaderCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

interface LoadingSpinnerProps {
  size?: string;
}

export const LoadingSpinner = ({ size = '1em' }: LoadingSpinnerProps) => {
  return (
    <div className={cn('flex flex-1 items-center justify-center text-inherit')}>
      <LoaderCircle className="animate-spin" size={size} />
    </div>
  );
};

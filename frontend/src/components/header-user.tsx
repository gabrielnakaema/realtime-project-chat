import { LogOut } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useAuth } from '@/hooks/use-auth';
import { cn } from '@/lib/utils';

export const HeaderUser = () => {
  const { user, logout } = useAuth();

  const name = user?.name ?? '';

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button type="button" className="flex items-center gap-2">
          <div
            className={cn(
              'text-foreground flex items-center justify-center rounded-full bg-blue-600 text-sm font-medium',
              'h-8 min-h-8 w-8 min-w-8 border border-transparent hover:outline-2 hover:outline-blue-200',
            )}
          >
            {name.charAt(0).toUpperCase()}
          </div>
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem asChild>
          <button type="button" onClick={logout} className="w-full">
            <LogOut className="h-4 w-4" />
            Logout
          </button>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

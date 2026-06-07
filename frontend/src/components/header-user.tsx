import { Link } from '@tanstack/react-router';
import { KeyRound, LogOut } from 'lucide-react';
import { useState } from 'react';
import { ChangePasswordDialog } from './change-password-dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useAuth } from '@/hooks/use-auth';
import { cn } from '@/lib/utils';

export const HeaderUser = () => {
  const { user, logout } = useAuth();
  const [isChangePasswordOpen, setIsChangePasswordOpen] = useState(false);

  const name = user?.name ?? '';

  return (
    <>
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
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => setIsChangePasswordOpen(true)}>
            <KeyRound className="h-4 w-4" />
            Change password
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <Link to="/mcp-access">
              <KeyRound className="h-4 w-4" />
              MCP Access
            </Link>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem asChild>
            <button type="button" onClick={logout} className="w-full">
              <LogOut className="h-4 w-4" />
              Logout
            </button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <ChangePasswordDialog open={isChangePasswordOpen} onOpenChange={setIsChangePasswordOpen} />
    </>
  );
};

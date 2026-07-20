import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { HeaderActions, HeaderLogo, HeaderShell, HeaderStart } from '@/shared/components/header-shell';

export const LandingHeader = () => (
  <HeaderShell width="contained" className="bg-background/90 sticky top-0 z-30 backdrop-blur-xl">
    <HeaderStart>
      <HeaderLogo />
      <span className="text-foreground text-sm font-semibold tracking-[-0.01em]">TaskFlow</span>
    </HeaderStart>

    <HeaderActions>
      <Link
        to="/login"
        className="text-muted-foreground hover:text-foreground hidden h-9 items-center px-3 text-sm font-medium transition-colors sm:inline-flex"
      >
        Log in
      </Link>
      <Link
        to="/sign-up"
        className="bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-9 items-center gap-2 rounded-md px-4 text-sm font-medium transition-colors"
      >
        Start a project
        <ArrowRight className="size-3.5" />
      </Link>
    </HeaderActions>
  </HeaderShell>
);

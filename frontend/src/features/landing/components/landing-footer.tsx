import { HeaderLogo } from '@/shared/components/header-shell';

export const LandingFooter = () => (
  <footer className="border-border border-t px-4 py-7 sm:px-6">
    <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 sm:flex-row">
      <div className="flex items-center gap-2">
        <HeaderLogo className="size-7" />
        <span className="text-foreground text-sm font-semibold">TaskFlow</span>
      </div>
      <p className="text-muted-foreground text-xs">Project planning, chat, and live updates in one place.</p>
    </div>
  </footer>
);

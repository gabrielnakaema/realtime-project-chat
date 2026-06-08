import * as React from 'react';
import { Command as CommandPrimitive } from 'cmdk';
import { Search } from 'lucide-react';

import { cn } from '@/lib/utils';

function Command({ className, ...props }: React.ComponentProps<typeof CommandPrimitive>) {
  return (
    <CommandPrimitive
      data-slot="command"
      className={cn(
        'flex h-full w-full flex-col overflow-hidden rounded-md bg-white text-slate-950 dark:bg-slate-950 dark:text-slate-50',
        className,
      )}
      {...props}
    />
  );
}

function CommandInput({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.Input>) {
  return (
    <div
      data-slot="command-input-wrapper"
      className="flex items-center border-b border-slate-200 px-3 dark:border-slate-700"
    >
      <Search className="mr-2 h-4 w-4 shrink-0 text-slate-500 dark:text-slate-400" />
      <CommandPrimitive.Input
        data-slot="command-input"
        className={cn(
          'flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-slate-500 disabled:cursor-not-allowed disabled:opacity-50 dark:placeholder:text-slate-400',
          className,
        )}
        {...props}
      />
    </div>
  );
}

function handleCommandListWheel(event: React.WheelEvent<HTMLDivElement>) {
  const element = event.currentTarget;
  if (element.scrollHeight <= element.clientHeight) {
    return;
  }

  const scrollTop = element.scrollTop;
  const scrollBottom = scrollTop + element.clientHeight;
  const scrollingUp = event.deltaY < 0;
  const scrollingDown = event.deltaY > 0;
  const canScrollUp = scrollTop > 0;
  const canScrollDown = scrollBottom < element.scrollHeight;

  if ((scrollingUp && canScrollUp) || (scrollingDown && canScrollDown)) {
    event.stopPropagation();
  }
}

function CommandList({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.List>) {
  return (
    <CommandPrimitive.List
      data-slot="command-list"
      className={cn('max-h-64 overflow-x-hidden overflow-y-auto overscroll-contain', className)}
      onWheel={handleCommandListWheel}
      onTouchMove={(event) => event.stopPropagation()}
      {...props}
    />
  );
}

function CommandEmpty({ ...props }: React.ComponentProps<typeof CommandPrimitive.Empty>) {
  return (
    <CommandPrimitive.Empty
      data-slot="command-empty"
      className="py-6 text-center text-sm text-slate-500 dark:text-slate-400"
      {...props}
    />
  );
}

function CommandGroup({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.Group>) {
  return (
    <CommandPrimitive.Group
      data-slot="command-group"
      className={cn(
        'overflow-hidden p-1 text-slate-950 dark:text-slate-50 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-slate-500 dark:[&_[cmdk-group-heading]]:text-slate-400',
        className,
      )}
      {...props}
    />
  );
}

function CommandItem({ className, ...props }: React.ComponentProps<typeof CommandPrimitive.Item>) {
  return (
    <CommandPrimitive.Item
      data-slot="command-item"
      className={cn(
        'relative flex cursor-default items-center gap-2 rounded-sm px-2 py-2 text-sm outline-none select-none data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50 data-[selected=true]:bg-slate-100 data-[selected=true]:text-slate-900 dark:data-[selected=true]:bg-slate-800 dark:data-[selected=true]:text-slate-100 [&_svg]:pointer-events-none [&_svg]:shrink-0',
        className,
      )}
      {...props}
    />
  );
}

export { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList };

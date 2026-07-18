import { useQuery } from '@tanstack/react-query';
import { Check } from 'lucide-react';
import { useRef, useState } from 'react';
import { useController, useFormContext } from 'react-hook-form';
import { LoadingSpinner } from '../loading';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '../ui/command';
import { Popover, PopoverAnchor, PopoverContent } from '../ui/popover';
import type { ITaskForm } from '@/schemas/task-schema';
import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { cn } from '@/lib/utils';
import { taskQueryKeys } from '@/services/query-keys';
import { suggestTaskCodes } from '@/services/tasks';
import { normalizeSearchQuery } from '@/utils/search';

const CODE_SUGGESTION_LIMIT = 8;
const CODE_SUGGESTION_DEBOUNCE_MS = 250;
const MIN_CODE_PREFIX_LENGTH = 2;

interface TaskCodeAutocompleteProps {
  projectId: string;
}

export const TaskCodeAutocomplete = ({ projectId }: TaskCodeAutocompleteProps) => {
  const { control } = useFormContext<ITaskForm>();
  const [open, setOpen] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const { field, fieldState } = useController({ control, name: 'code' });
  const codeValue = field.value ?? '';

  const prefix = normalizeSearchQuery(codeValue);
  const debouncedPrefix = useDebouncedValue(prefix, CODE_SUGGESTION_DEBOUNCE_MS);
  const params = {
    projectId,
    prefix: debouncedPrefix,
    limit: CODE_SUGGESTION_LIMIT,
  };

  const {
    data: suggestions = [],
    isError,
    isFetching,
    refetch,
  } = useQuery({
    queryKey: taskQueryKeys.codeSuggestions(params),
    queryFn: () => suggestTaskCodes(params),
    enabled: open && debouncedPrefix.length >= MIN_CODE_PREFIX_LENGTH,
    staleTime: 10_000,
  });

  const hasEnoughPrefix = debouncedPrefix.length >= MIN_CODE_PREFIX_LENGTH;
  const isDebouncing = prefix !== debouncedPrefix;

  return (
    <div className="w-full space-y-2">
      <span className="text-foreground block text-sm font-medium">Code</span>
      <Command
        label="Code"
        shouldFilter={false}
        className="border-border bg-card focus-within:ring-ring h-auto overflow-visible border focus-within:border-transparent focus-within:ring-2 [&_[data-slot=command-input-wrapper]]:border-b-0"
      >
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverAnchor className="block w-full">
            <CommandInput
              name={field.name}
              ref={(element) => {
                inputRef.current = element;
                field.ref(element);
              }}
              placeholder="TASK-101"
              value={codeValue}
              onValueChange={(value) => {
                field.onChange(value);
                setOpen(true);
              }}
              onBlur={field.onBlur}
              onFocus={() => setOpen(true)}
              onKeyDown={(event) => {
                if (event.key === 'ArrowDown') {
                  setOpen(true);
                }

                if (event.key === 'Escape') {
                  setOpen(false);
                }
              }}
              aria-expanded={open}
              aria-controls="code-suggestions"
              aria-invalid={Boolean(fieldState.error)}
              aria-describedby={fieldState.error ? 'code-error' : undefined}
            />
          </PopoverAnchor>
          <PopoverContent
            className="pointer-events-auto z-[60] w-[var(--radix-popover-trigger-width)] p-0"
            align="start"
            onOpenAutoFocus={(e) => e.preventDefault()}
            onCloseAutoFocus={(e) => e.preventDefault()}
            onMouseDown={(e) => e.preventDefault()}
            onInteractOutside={(event) => {
              if (event.target instanceof Node && inputRef.current?.contains(event.target)) {
                event.preventDefault();
              }
            }}
          >
            <CommandList id="code-suggestions">
              {!isDebouncing && !hasEnoughPrefix && (
                <div className="text-muted-foreground py-6 text-center text-sm">Type at least 2 characters.</div>
              )}
              {(isDebouncing || (hasEnoughPrefix && isFetching)) && (
                <div className="flex items-center justify-center py-6">
                  <LoadingSpinner size="1.5rem" />
                </div>
              )}
              {!isDebouncing && hasEnoughPrefix && !isFetching && isError && (
                <div className="text-muted-foreground space-y-3 px-4 py-6 text-center text-sm">
                  <p>Could not load code suggestions.</p>
                  <button type="button" className="text-primary font-medium hover:underline" onClick={() => refetch()}>
                    Try again
                  </button>
                </div>
              )}
              {!isDebouncing && hasEnoughPrefix && !isFetching && !isError && suggestions.length === 0 && (
                <CommandEmpty>No matching codes found.</CommandEmpty>
              )}
              {!isDebouncing && hasEnoughPrefix && !isFetching && !isError && suggestions.length > 0 && (
                <CommandGroup heading="Code suggestions">
                  {suggestions.map((suggestion) => {
                    const selected = suggestion.code === codeValue;

                    return (
                      <CommandItem
                        key={`${suggestion.kind}-${suggestion.code}`}
                        value={suggestion.code}
                        onSelect={() => {
                          field.onChange(suggestion.code);
                          setOpen(false);
                        }}
                        className="cursor-pointer"
                      >
                        <Check className={cn('h-4 w-4', selected ? 'opacity-100' : 'opacity-0')} />
                        <div className="min-w-0 flex-1">
                          <div className="truncate font-medium">{suggestion.code}</div>
                          <p className="text-muted-foreground truncate text-xs">
                            {suggestion.kind === 'next' ? 'Next available code' : 'Existing task code'}
                          </p>
                        </div>
                      </CommandItem>
                    );
                  })}
                </CommandGroup>
              )}
            </CommandList>
          </PopoverContent>
        </Popover>
      </Command>
      {fieldState.error?.message && (
        <p id="code-error" className="text-destructive text-sm">
          {fieldState.error.message}
        </p>
      )}
    </div>
  );
};

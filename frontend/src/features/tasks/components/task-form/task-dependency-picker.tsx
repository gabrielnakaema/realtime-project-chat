import { useQuery } from '@tanstack/react-query';
import { Check, ChevronsUpDown, X } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import { TaskCodeBadge } from '../task-code-badge';
import { formatTaskDependencyLabel } from './task-form-utils';
import type { ITaskForm } from '@/features/tasks/schemas/task.schema';
import type { TaskDependencyRef } from '@/features/tasks/types/task';
import { Button } from '@/components/button';
import { LoadingSpinner } from '@/shared/components/loading';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/shared/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/components/ui/popover';
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value';
import { searchProjectTasksForDependencies } from '@/features/tasks/services/tasks';
import { taskQueryKeys } from '@/shared/services/query-keys';
import { normalizeSearchQuery } from '@/shared/utils/search';
import { cn } from '@/lib/utils';

const SEARCH_LIMIT = 20;
const SEARCH_DEBOUNCE_MS = 300;

interface TaskDependencyPickerProps {
  projectId: string;
  excludeTaskId?: string;
  initialSelectedTasks?: TaskDependencyRef[];
}

export const TaskDependencyPicker = ({
  projectId,
  excludeTaskId,
  initialSelectedTasks = [],
}: TaskDependencyPickerProps) => {
  const {
    control,
    formState: { errors },
  } = useFormContext<ITaskForm>();
  const [open, setOpen] = useState(false);
  const [searchValue, setSearchValue] = useState('');
  const debouncedSearch = useDebouncedValue(normalizeSearchQuery(searchValue), SEARCH_DEBOUNCE_MS);

  const params = useMemo(
    () => ({
      projectId,
      query: debouncedSearch,
      excludeTaskId,
      limit: SEARCH_LIMIT,
    }),
    [projectId, debouncedSearch, excludeTaskId],
  );

  const { data: searchResults = [], isFetching } = useQuery({
    queryKey: taskQueryKeys.searchProjectDependencies(params),
    queryFn: () => searchProjectTasksForDependencies(params),
    enabled: open && debouncedSearch.length > 0,
    staleTime: 10_000,
  });

  const [pickedTasksById, setPickedTasksById] = useState(() => new Map<string, TaskDependencyRef>());

  const knownTasksById = useMemo(() => {
    const map = new Map<string, TaskDependencyRef>();

    for (const task of initialSelectedTasks) {
      map.set(task.id, task);
    }

    for (const task of searchResults) {
      map.set(task.id, task);
    }

    for (const [taskId, task] of pickedTasksById) {
      map.set(taskId, task);
    }

    return map;
  }, [initialSelectedTasks, searchResults, pickedTasksById]);

  return (
    <Controller
      control={control}
      name="depends_on_task_ids"
      render={({ field }) => {
        const selectedTaskIds = field.value;

        const selectedTasks = selectedTaskIds
          .map((taskId) => knownTasksById.get(taskId))
          .filter((task): task is TaskDependencyRef => Boolean(task));

        const toggleTask = (task: TaskDependencyRef) => {
          const nextTaskIds = selectedTaskIds.includes(task.id)
            ? selectedTaskIds.filter((value) => value !== task.id)
            : [...selectedTaskIds, task.id];

          setPickedTasksById((current) => {
            const next = new Map(current);
            next.set(task.id, task);
            return next;
          });
          field.onChange(nextTaskIds);
        };

        return (
          <div className="space-y-3">
            {selectedTasks.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {selectedTasks.map((task) => (
                  <div
                    key={task.id}
                    className="border-border bg-muted text-foreground inline-flex max-w-full min-w-0 items-center gap-2 rounded-full border px-2.5 py-1 text-xs font-medium"
                  >
                    {task.code ? <TaskCodeBadge code={task.code} /> : null}
                    <span className="truncate">{task.title}</span>
                    <button
                      type="button"
                      aria-label={`Remove dependency ${task.title}`}
                      className="text-muted-foreground hover:bg-accent hover:text-foreground rounded-full p-0.5 transition-colors"
                      onClick={() => toggleTask(task)}
                    >
                      <X className="h-3.5 w-3.5" />
                    </button>
                  </div>
                ))}
              </div>
            )}

            <Popover open={open} onOpenChange={setOpen} modal>
              <PopoverTrigger asChild>
                <Button
                  type="button"
                  variant="secondary"
                  className="border-border bg-card text-foreground hover:bg-muted w-full justify-between border"
                >
                  <span>{selectedTasks.length > 0 ? 'Add another dependency' : 'Search tasks to depend on'}</span>
                  <ChevronsUpDown className="h-4 w-4 shrink-0 opacity-60" />
                </Button>
              </PopoverTrigger>
              <PopoverContent
                className="pointer-events-auto z-[60] w-[min(24rem,calc(100vw-2rem))] p-0"
                align="start"
                onWheel={(event) => event.stopPropagation()}
                onTouchMove={(event) => event.stopPropagation()}
              >
                <Command shouldFilter={false}>
                  <CommandInput
                    placeholder="Search by code, title, id, or description..."
                    value={searchValue}
                    onValueChange={setSearchValue}
                  />
                  <CommandList>
                    {debouncedSearch.length === 0 && (
                      <div className="text-muted-foreground py-6 text-center text-sm">
                        Type to search project tasks.
                      </div>
                    )}
                    {debouncedSearch.length > 0 && isFetching && (
                      <div className="flex items-center justify-center py-6">
                        <LoadingSpinner size="1.5rem" />
                      </div>
                    )}
                    {debouncedSearch.length > 0 && !isFetching && searchResults.length === 0 && (
                      <CommandEmpty>No matching tasks found.</CommandEmpty>
                    )}
                    {debouncedSearch.length > 0 && !isFetching && searchResults.length > 0 && (
                      <CommandGroup heading="Tasks">
                        {searchResults.map((task) => {
                          const selected = selectedTaskIds.includes(task.id);

                          return (
                            <CommandItem
                              key={task.id}
                              value={task.id}
                              onSelect={() => toggleTask(task)}
                              className="cursor-pointer"
                            >
                              <Check className={cn('h-4 w-4', selected ? 'opacity-100' : 'opacity-0')} />
                              <div className="min-w-0 flex-1">
                                <div className="flex min-w-0 items-center gap-2">
                                  {task.code ? <TaskCodeBadge code={task.code} /> : null}
                                  <span className="truncate font-medium">{task.title}</span>
                                </div>
                                <p className="text-muted-foreground truncate text-xs">
                                  {formatTaskDependencyLabel(task)}
                                </p>
                              </div>
                            </CommandItem>
                          );
                        })}
                      </CommandGroup>
                    )}
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>

            {errors.depends_on_task_ids?.message && (
              <p className="text-destructive text-sm">{errors.depends_on_task_ids.message}</p>
            )}
          </div>
        );
      }}
    />
  );
};

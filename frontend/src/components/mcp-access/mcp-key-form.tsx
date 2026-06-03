import { zodResolver } from '@hookform/resolvers/zod';
import { Controller, useForm } from 'react-hook-form';
import { useEffect } from 'react';
import { Button } from '../button';
import { Input } from '../input';
import { LoadingSpinner } from '../loading';
import { DialogFooter } from '../ui/dialog';
import type { IMCPAPIKeyForm } from '@/schemas/mcp-api-key-schema';
import { useMCPAPIAvailableScopes } from '@/hooks/use-mcp-api-available-scopes';
import { mcpAPIKeySchema } from '@/schemas/mcp-api-key-schema';
import { cn } from '@/lib/utils';

interface MCPAPIKeyFormProps {
  defaultValues: IMCPAPIKeyForm;
  errorMessage?: string | null;
  isSubmitting: boolean;
  submitLabel: string;
  onCancel: () => void;
  onFormChange?: () => void;
  onSubmit: (values: IMCPAPIKeyForm) => void;
}

export const MCPAPIKeyForm = ({
  defaultValues,
  errorMessage,
  isSubmitting,
  submitLabel,
  onCancel,
  onFormChange,
  onSubmit,
}: MCPAPIKeyFormProps) => {
  const { availableScopes, isLoadingScopes, isScopeLoadFailed, refetchScopes } = useMCPAPIAvailableScopes();
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<IMCPAPIKeyForm>({
    resolver: zodResolver(mcpAPIKeySchema),
    defaultValues,
    mode: 'onChange',
  });

  useEffect(() => {
    reset(defaultValues);
  }, [defaultValues, reset]);

  if (isLoadingScopes) {
    return (
      <>
        <div className="flex min-h-72 flex-1 items-center justify-center px-6 py-5">
          <LoadingSpinner size="2rem" />
        </div>

        <DialogFooter className="border-t border-slate-200 bg-slate-50/80 px-6 py-4 dark:border-slate-800 dark:bg-slate-900/80">
          <Button type="button" variant="secondary" onClick={onCancel}>
            Cancel
          </Button>
        </DialogFooter>
      </>
    );
  }

  if (isScopeLoadFailed) {
    return (
      <>
        <div className="px-6 py-5">
          <div className="space-y-3 rounded-2xl border border-red-200 bg-red-50 p-5 dark:border-red-900/80 dark:bg-red-950/30">
            <div className="space-y-1">
              <p className="font-semibold text-red-900 dark:text-red-100">Could not load available scopes</p>
              <p className="text-sm text-red-700 dark:text-red-300">
                Retry to load the current MCP permissions before saving this key.
              </p>
            </div>
            <Button type="button" variant="secondary" onClick={() => void refetchScopes()}>
              Retry
            </Button>
          </div>
        </div>

        <DialogFooter className="border-t border-slate-200 bg-slate-50/80 px-6 py-4 dark:border-slate-800 dark:bg-slate-900/80">
          <Button type="button" variant="secondary" onClick={onCancel}>
            Cancel
          </Button>
        </DialogFooter>
      </>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-1 flex-col overflow-hidden">
      <div className="flex-1 overflow-y-auto px-6 py-5">
        <section className="space-y-4">
          <Input
            id="mcp-key-name"
            label="Key name"
            placeholder="Example: Claude desktop"
            error={errors.name?.message}
            {...register('name', {
              onChange: () => onFormChange?.(),
            })}
          />

          <div className="space-y-2">
            <div>
              <p className="text-sm font-medium text-slate-700 dark:text-slate-300">Scopes</p>
              <p className="text-sm text-slate-600 dark:text-slate-400">
                Select one or more permissions for this MCP client.
              </p>
            </div>

            <Controller
              control={control}
              name="scopes"
              render={({ field }) => {
                const selectedScopes = field.value;

                return (
                  <div className="grid gap-3">
                    {availableScopes.map((scope) => {
                      const selected = selectedScopes.includes(scope.scope);

                      return (
                        <label
                          key={scope.scope}
                          className={cn(
                            'flex cursor-pointer items-start gap-3 rounded-xl border p-3 transition-colors',
                            selected
                              ? 'border-blue-300 bg-blue-50 dark:border-blue-800 dark:bg-blue-950/30'
                              : 'border-slate-200 bg-slate-50/70 hover:bg-slate-100 dark:border-slate-700 dark:bg-slate-900/60 dark:hover:bg-slate-900',
                          )}
                        >
                          <input
                            type="checkbox"
                            checked={selected}
                            onChange={() => {
                              onFormChange?.();
                              const nextScopes = selected
                                ? selectedScopes.filter((value) => value !== scope.scope)
                                : [...selectedScopes, scope.scope];
                              field.onChange(nextScopes);
                            }}
                            className="mt-1 h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-2 focus:ring-blue-500"
                          />
                          <div className="space-y-1">
                            <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{scope.label}</p>
                            <p className="text-sm text-slate-600 dark:text-slate-400">{scope.title}</p>
                          </div>
                        </label>
                      );
                    })}
                  </div>
                );
              }}
            />

            {errors.scopes?.message && <p className="text-sm text-red-500">{errors.scopes.message}</p>}
            {errorMessage && <p className="text-sm text-red-500">{errorMessage}</p>}
          </div>
        </section>
      </div>

      <DialogFooter className="border-t border-slate-200 bg-slate-50/80 px-6 py-4 dark:border-slate-800 dark:bg-slate-900/80">
        <Button type="button" variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? <LoadingSpinner size="1.25rem" /> : submitLabel}
        </Button>
      </DialogFooter>
    </form>
  );
};

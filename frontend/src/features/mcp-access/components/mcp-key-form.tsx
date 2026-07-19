import { zodResolver } from '@hookform/resolvers/zod';
import { Controller, useForm } from 'react-hook-form';
import { useEffect } from 'react';
import { Button } from '../../../components/button';
import { Input } from '../../../components/input';
import { DialogFooter } from '../../../shared/components/ui/dialog';
import type { IMCPAPIKeyForm } from '@/features/mcp-access/schemas/mcp-api-key.schema';
import { LoadingSpinner } from '@/shared/components/loading';
import { useMCPAPIAvailableScopes } from '@/features/mcp-access/hooks/use-mcp-api-available-scopes';
import { mcpAPIKeySchema } from '@/features/mcp-access/schemas/mcp-api-key.schema';
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

        <DialogFooter className="border-border bg-muted/80 border-t px-6 py-4">
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
          <div className="border-destructive/30 bg-destructive/10 space-y-3 rounded-2xl border p-5">
            <div className="space-y-1">
              <p className="text-destructive font-semibold">Could not load available scopes</p>
              <p className="text-destructive text-sm">
                Retry to load the current MCP permissions before saving this key.
              </p>
            </div>
            <Button type="button" variant="secondary" onClick={() => void refetchScopes()}>
              Retry
            </Button>
          </div>
        </div>

        <DialogFooter className="border-border bg-muted/80 border-t px-6 py-4">
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
              <p className="text-foreground text-sm font-medium">Scopes</p>
              <p className="text-muted-foreground text-sm">Select one or more permissions for this MCP client.</p>
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
                            selected ? 'border-primary/30 bg-primary/10' : 'border-border bg-muted/70 hover:bg-muted',
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
                            className="border-border text-primary focus:ring-ring mt-1 h-4 w-4 rounded focus:ring-2"
                          />
                          <div className="space-y-1">
                            <p className="text-foreground text-sm font-medium">{scope.label}</p>
                            <p className="text-muted-foreground text-sm">{scope.title}</p>
                          </div>
                        </label>
                      );
                    })}
                  </div>
                );
              }}
            />

            {errors.scopes?.message && <p className="text-destructive text-sm">{errors.scopes.message}</p>}
            {errorMessage && <p className="text-destructive text-sm">{errorMessage}</p>}
          </div>
        </section>
      </div>

      <DialogFooter className="border-border bg-muted/80 border-t px-6 py-4">
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

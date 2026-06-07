import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { Button } from './button';
import { Input } from './input';
import { LoadingSpinner } from './loading';
import { Textarea } from './textarea';
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle } from './ui/dialog';
import type { IEditColumnForm } from '@/schemas/edit-column-schema';
import type { Column } from '@/types/board';
import { editColumnSchema } from '@/schemas/edit-column-schema';
import { invalidateProjectBoardData } from '@/services/project-board-invalidation';
import { updateProjectColumn } from '@/services/projects';
import { handleSuccess } from '@/utils/handle-success';

interface EditColumnDialogProps {
  column: Column;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const EditColumnDialog = ({ column, open, onOpenChange }: EditColumnDialogProps) => {
  const queryClient = useQueryClient();

  const {
    handleSubmit,
    register,
    reset,
    watch,
    formState: { errors },
  } = useForm<IEditColumnForm>({
    resolver: zodResolver(editColumnSchema),
    defaultValues: {
      name: column.title,
      description: column.description,
      color: column.color,
      is_done_column: column.isDoneColumn,
    },
  });

  const color = watch('color');
  const isDoneColumn = watch('is_done_column');

  useEffect(() => {
    if (!open) {
      return;
    }

    reset({
      name: column.title,
      description: column.description,
      color: column.color,
      is_done_column: column.isDoneColumn,
    });
  }, [column.color, column.description, column.isDoneColumn, column.title, open, reset]);

  const { mutate, isPending } = useMutation({
    mutationFn: updateProjectColumn,
    onSuccess: async () => {
      await invalidateProjectBoardData(queryClient);
      handleSuccess('Column updated successfully');
      onOpenChange(false);
    },
  });

  const onSubmit = handleSubmit((values) => {
    mutate({
      projectId: column.project_id,
      columnId: column.columnId,
      name: values.name,
      description: values.description,
      color: values.color,
      is_done_column: values.is_done_column,
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit column</DialogTitle>
          <DialogDescription>
            Update the column details and refresh the board with the latest project and task data.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={onSubmit} className="space-y-4">
          <Input
            id={`column-name-${column.id}`}
            label="Column name"
            error={errors.name?.message}
            {...register('name')}
          />

          <Textarea
            id={`column-description-${column.id}`}
            label="Column description"
            rows={4}
            placeholder="Optional guidance or instructions for this column."
            error={errors.description?.message}
            {...register('description')}
          />

          <div className="space-y-2">
            <label
              htmlFor={`column-color-${column.id}`}
              className="block text-sm font-medium text-slate-700 dark:text-slate-300"
            >
              Color
            </label>
            <div className="flex items-center gap-3 rounded-xl border border-slate-200 px-3 py-2 dark:border-slate-700">
              <input
                id={`column-color-${column.id}`}
                type="color"
                className="h-10 w-12 cursor-pointer rounded border-0 bg-transparent p-0"
                {...register('color')}
              />
              <span className="font-mono text-sm text-slate-700 dark:text-slate-200">{color}</span>
            </div>
            {errors.color?.message && <p className="text-sm text-red-500">{errors.color.message}</p>}
          </div>

          <label className="flex items-center gap-3 rounded-xl border border-slate-200 px-3 py-3 dark:border-slate-700">
            <input type="checkbox" className="h-4 w-4" {...register('is_done_column')} />
            <div className="space-y-1">
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">Done column</p>
              <p className="text-sm text-slate-600 dark:text-slate-400">
                {isDoneColumn
                  ? 'Tasks moved here will be treated as completed.'
                  : 'Mark this column as the completed stage for the board.'}
              </p>
            </div>
          </label>

          <div className="flex items-center justify-end gap-3 border-t border-slate-200 pt-4 dark:border-slate-700">
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                Cancel
              </Button>
            </DialogClose>
            <Button type="submit" disabled={isPending}>
              {isPending ? <LoadingSpinner size="1.5em" /> : 'Save changes'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};

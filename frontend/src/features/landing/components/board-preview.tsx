import { useCallback, useState } from 'react';
import { PreviewBoardColumn } from './preview-board-column';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { PreviewColumnId } from '@/features/landing/data/preview-board';
import { initialPreviewBoard, previewColumns, previewMemberInitials } from '@/features/landing/data/preview-board';
import { movePreviewTask } from '@/features/landing/data/move-preview-task';
import { HeaderLogo } from '@/shared/components/header-shell';

interface LastMove {
  code: string;
  columnName: string;
}

export const BoardPreview = () => {
  const [board, setBoard] = useState(initialPreviewBoard);
  const [lastMove, setLastMove] = useState<LastMove | null>(null);

  const handleTaskDrop = useCallback(
    (destinationColumnId: PreviewColumnId) =>
      (edge: Edge | null, droppedTaskId: string, targetTaskId: string | null) => {
        setBoard((current) => {
          const next = movePreviewTask({
            board: current,
            taskId: droppedTaskId,
            destinationColumnId,
            targetTaskId,
            edge,
          });

          if (next !== current) {
            const movedTask = Object.values(current)
              .flat()
              .find((task) => task.id === droppedTaskId);
            const columnName = previewColumns.find((column) => column.id === destinationColumnId)?.name ?? '';

            if (movedTask) {
              setLastMove({ code: movedTask.code, columnName });
            }
          }

          return next;
        });
      },
    [],
  );

  return (
    <section id="product-preview" className="scroll-mt-24 px-4 pt-6 pb-20 sm:px-6 lg:pb-28">
      <div className="mx-auto max-w-7xl">
        <div className="mb-7 flex flex-col justify-between gap-4 px-1 sm:flex-row sm:items-end">
          <div>
            <p className="text-primary mb-3 text-xs font-semibold tracking-[0.16em] uppercase">Interactive preview</p>
            <h2 className="text-foreground text-2xl font-semibold tracking-[-0.025em] sm:text-3xl">
              See the whole launch in motion.
            </h2>
          </div>
          <p className="text-muted-foreground max-w-md text-sm leading-6">
            Drag any task between columns and watch the board update—just like it does for every teammate.
          </p>
        </div>

        <div className="border-border bg-card overflow-hidden rounded-2xl border shadow-2xl shadow-black/20">
          <div className="border-border flex h-16 items-center gap-3 border-b px-4 sm:px-5">
            <HeaderLogo className="size-7" />
            <div className="bg-border hidden h-5 w-px sm:block" />
            <div className="min-w-0">
              <p className="text-foreground truncate text-sm font-semibold">Product launch</p>
            </div>
            <div className="ml-auto hidden -space-x-1.5 sm:flex" aria-label="Project members">
              {previewMemberInitials.map((initials) => (
                <span
                  key={initials}
                  className="border-card bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-full border-2 text-[9px] font-semibold"
                >
                  {initials}
                </span>
              ))}
            </div>
          </div>

          <div className="bg-background overflow-x-auto p-4 sm:p-5">
            <div className="grid min-w-[920px] grid-cols-4 gap-4">
              {previewColumns.map((column) => (
                <PreviewBoardColumn
                  key={column.id}
                  column={column}
                  tasks={board[column.id]}
                  onTaskDrop={handleTaskDrop(column.id)}
                />
              ))}
            </div>
          </div>

          <div
            className="border-border bg-card flex items-center gap-2 border-t px-4 py-3 text-xs sm:px-5"
            role="status"
            aria-live="polite"
          >
            {lastMove ? (
              <>
                <span className="bg-success size-2 rounded-full shadow-[0_0_0_4px_color-mix(in_srgb,var(--success)_12%,transparent)]" />
                <span className="text-muted-foreground">
                  <strong className="text-foreground font-medium">You</strong> moved {lastMove.code} to{' '}
                  <strong className="text-foreground font-medium">{lastMove.columnName}</strong>
                </span>
              </>
            ) : (
              <>
                <span className="bg-muted-foreground/50 size-2 rounded-full" />
                <span className="text-muted-foreground">Drag a card to move it across the board.</span>
              </>
            )}
          </div>
        </div>
      </div>
    </section>
  );
};

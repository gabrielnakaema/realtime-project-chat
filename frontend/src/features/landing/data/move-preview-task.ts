import { previewColumns } from './preview-board';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { PreviewBoardState, PreviewColumnId, PreviewTask } from './preview-board';

type MutableBoardState = Record<PreviewColumnId, PreviewTask[]>;

const emptyColumns = (): MutableBoardState =>
  previewColumns.reduce((acc, column) => {
    acc[column.id] = [];
    return acc;
  }, {} as MutableBoardState);

interface MovePreviewTaskArgs {
  board: PreviewBoardState;
  taskId: string;
  destinationColumnId: PreviewColumnId;
  targetTaskId: string | null;
  edge: Edge | null;
}

export const movePreviewTask = ({
  board,
  taskId,
  destinationColumnId,
  targetTaskId,
  edge,
}: MovePreviewTaskArgs): PreviewBoardState => {
  if (taskId === targetTaskId) {
    return board;
  }

  const next = emptyColumns();
  let moved: PreviewTask | undefined;

  for (const column of previewColumns) {
    next[column.id] = board[column.id].filter((task) => {
      if (task.id === taskId) {
        moved = task;
        return false;
      }
      return true;
    });
  }

  if (!moved) {
    return board;
  }

  const destination = next[destinationColumnId];
  const targetIndex = targetTaskId ? destination.findIndex((task) => task.id === targetTaskId) : -1;

  if (targetIndex === -1) {
    destination.push(moved);
  } else {
    destination.splice(edge === 'bottom' ? targetIndex + 1 : targetIndex, 0, moved);
  }

  return next;
};

import { describe, expect, it } from 'vitest';
import { movePreviewTask } from './move-preview-task';
import type { PreviewBoardState } from './preview-board';

const board: PreviewBoardState = {
  backlog: [
    { id: 'a', code: 'A', title: 'A', priority: 'Low', owner: 'AA', ownerName: 'A A', tags: [] },
    { id: 'b', code: 'B', title: 'B', priority: 'Low', owner: 'BB', ownerName: 'B B', tags: [] },
  ],
  doing: [{ id: 'c', code: 'C', title: 'C', priority: 'Low', owner: 'CC', ownerName: 'C C', tags: [] }],
  review: [],
  done: [],
};

const ids = (tasks: PreviewBoardState[keyof PreviewBoardState]) => tasks.map((task) => task.id);

describe('movePreviewTask', () => {
  it('reorders a task within its column using the drop edge', () => {
    const next = movePreviewTask({
      board,
      taskId: 'b',
      destinationColumnId: 'backlog',
      targetTaskId: 'a',
      edge: 'top',
    });

    expect(ids(next.backlog)).toEqual(['b', 'a']);
  });

  it('moves a task to another column above the target card', () => {
    const next = movePreviewTask({
      board,
      taskId: 'c',
      destinationColumnId: 'backlog',
      targetTaskId: 'b',
      edge: 'bottom',
    });

    expect(ids(next.backlog)).toEqual(['a', 'b', 'c']);
    expect(ids(next.doing)).toEqual([]);
  });

  it('appends to the end when dropped on empty column space', () => {
    const next = movePreviewTask({ board, taskId: 'a', destinationColumnId: 'review', targetTaskId: null, edge: null });

    expect(ids(next.review)).toEqual(['a']);
    expect(ids(next.backlog)).toEqual(['b']);
  });

  it('returns the same board when a task is dropped onto itself', () => {
    const next = movePreviewTask({
      board,
      taskId: 'a',
      destinationColumnId: 'backlog',
      targetTaskId: 'a',
      edge: 'top',
    });

    expect(next).toBe(board);
  });
});

// @vitest-environment jsdom

import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { TaskCard } from './task-card';
import type { Task } from '@/types/task';
import { useTaskDetailsRouting } from '@/hooks/use-task-details-routing';

vi.mock('@/hooks/use-task-details-routing', () => ({
  useTaskDetailsRouting: vi.fn(),
}));

vi.mock('@atlaskit/pragmatic-drag-and-drop/combine', () => ({
  combine:
    (...cleanups: Array<() => void>) =>
    () => {
      cleanups.forEach((cleanup) => cleanup());
    },
}));

vi.mock('@atlaskit/pragmatic-drag-and-drop/element/adapter', () => ({
  draggable: () => () => {},
  dropTargetForElements: () => () => {},
}));

vi.mock('@atlaskit/pragmatic-drag-and-drop-react-drop-indicator/box', () => ({
  DropIndicator: () => null,
}));

vi.mock('./avatar', () => ({
  Avatar: ({ name }: { name: string }) => <div>{name}</div>,
}));

const mockUseTaskDetailsRouting = vi.mocked(useTaskDetailsRouting);

const task: Task = {
  id: 'task-1',
  project_id: 'project-1',
  title: 'Ship prop-drilling fix',
  description: '<p>Refactor the board click path</p>',
  code: 'TASK-1',
  status: 'todo',
  project_column_id: 'column-1',
  project_column: null,
  created_at: '2026-06-05T10:00:00.000Z',
  updated_at: '2026-06-05T10:00:00.000Z',
  priority: 'high',
  order: '1',
  version: 0,
  responsible_id: null,
  due_date: null,
  done_at: null,
  archived_at: null,
  tags: ['frontend'],
  author_id: 'user-1',
  author: {
    id: 'user-1',
    name: 'Alex',
    email: 'alex@example.com',
  },
  responsible: null,
  updates: [],
  project: null,
};

afterEach(() => {
  document.body.innerHTML = '';
});

beforeEach(() => {
  mockUseTaskDetailsRouting.mockReset();
});

describe('TaskCard', () => {
  it('opens task details when the card is clicked', () => {
    const openTask = vi.fn();
    mockUseTaskDetailsRouting.mockReturnValue({
      selectedTaskId: '',
      selectedCommentId: undefined,
      selectedCommentCreatedAt: undefined,
      isEditingTask: false,
      openTask,
      closeTask: vi.fn(),
      startEditingTask: vi.fn(),
      stopEditingTask: vi.fn(),
    });

    render(
      <TaskCard
        task={task}
        onDrop={vi.fn()}
        columnSurface={{ backgroundColor: '', accentColor: '', badgeBackground: '', borderColor: '' }}
      />,
    );

    fireEvent.click(screen.getByText('Ship prop-drilling fix'));

    expect(openTask).toHaveBeenCalledWith('task-1');
  });
});

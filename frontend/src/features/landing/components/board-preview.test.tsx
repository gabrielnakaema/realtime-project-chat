// @vitest-environment jsdom

import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { BoardPreview } from './board-preview';

vi.mock('@atlaskit/pragmatic-drag-and-drop/combine', () => ({
  combine:
    (...cleanups: Array<() => void>) =>
    () =>
      cleanups.forEach((fn) => fn()),
}));

vi.mock('@atlaskit/pragmatic-drag-and-drop/element/adapter', () => ({
  draggable: () => () => {},
  dropTargetForElements: () => () => {},
}));

vi.mock('@atlaskit/pragmatic-drag-and-drop-react-drop-indicator/box', () => ({
  DropIndicator: () => null,
}));

afterEach(() => {
  cleanup();
});

describe('BoardPreview', () => {
  it('renders every preview column with its tasks', () => {
    render(<BoardPreview />);

    expect(screen.getByText('Backlog')).toBeTruthy();
    expect(screen.getByText('Done')).toBeTruthy();
    expect(screen.getByText('Ship the new landing page')).toBeTruthy();
    expect(screen.getByText('Connect realtime activity')).toBeTruthy();
  });

  it('prompts the user to drag a card before any move happens', () => {
    render(<BoardPreview />);

    expect(screen.getByRole('status').textContent).toContain('Drag a card to move it across the board.');
  });
});

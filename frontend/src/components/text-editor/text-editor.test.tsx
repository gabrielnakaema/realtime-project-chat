// @vitest-environment jsdom

import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { TextEditor } from '.';

beforeAll(() => {
  Object.defineProperty(Node.prototype, 'getBoundingClientRect', {
    configurable: true,
    value: () => new DOMRect(),
  });
  Object.defineProperty(Range.prototype, 'getBoundingClientRect', {
    configurable: true,
    value: () => new DOMRect(),
  });
});

afterEach(() => {
  document.body.innerHTML = '';
});

const renderEditor = (initialValue = '') => {
  const onChange = vi.fn();
  const result = render(
    <TextEditor initialValue={initialValue} onChange={onChange} label="Description" id="description" />,
  );
  return { ...result, onChange };
};

describe('TextEditor lists', () => {
  it.each([
    { button: 'Bulleted List', selector: 'ul', otherSelector: 'ol' },
    { button: 'Numbered List', selector: 'ol', otherSelector: 'ul' },
  ])('creates and toggles a $button without flattening its content', async ({ button, selector, otherSelector }) => {
    const { container, onChange } = renderEditor('<p>Keep me</p>');

    fireEvent.click(screen.getByRole('button', { name: button }));

    await waitFor(() => expect(container.querySelector(`${selector} > li`)).toHaveProperty('textContent', 'Keep me'));
    expect(container.querySelector(otherSelector)).toBeNull();
    expect(onChange).toHaveBeenLastCalledWith(expect.stringContaining(`<${selector}>`));
    expect(onChange).toHaveBeenLastCalledWith(expect.stringContaining('<li'));
    await waitFor(() => expect(screen.getByRole('button', { name: button }).getAttribute('aria-pressed')).toBe('true'));

    fireEvent.click(screen.getByRole('button', { name: button }));

    await waitFor(() => expect(container.querySelector(selector)).toBeNull());
    expect(container.querySelector('[contenteditable="true"] > p')).toHaveProperty('textContent', 'Keep me');
    await waitFor(() => expect(onChange).toHaveBeenLastCalledWith(expect.stringContaining('<p')));
  });

  it('round-trips existing ordered and unordered lists while preserving surrounding prose', () => {
    const initialValue =
      '<p>Before</p><ul><li>Bullet one</li><li>Bullet two</li></ul><ol><li>First</li><li>Second</li></ol><p>After</p>';
    const { container } = renderEditor(initialValue);
    const editable = container.querySelector<HTMLElement>('[contenteditable="true"]');

    expect(editable?.querySelectorAll(':scope > p')).toHaveLength(2);
    expect(editable?.querySelectorAll(':scope > ul > li')).toHaveLength(2);
    expect(editable?.querySelectorAll(':scope > ol > li')).toHaveLength(2);
    expect(editable?.textContent).toBe('BeforeBullet oneBullet twoFirstSecondAfter');
  });

  it('applies visible list marker styles to the editable content', () => {
    const { container } = renderEditor('<ul><li>Visible marker</li></ul>');
    const editable = container.querySelector('[contenteditable="true"]');

    expect(Array.from(editable?.classList ?? [])).toEqual(
      expect.arrayContaining(['[&_ul]:list-disc', '[&_ol]:list-decimal', '[&_ul]:pl-6', '[&_ol]:pl-6']),
    );
  });
});

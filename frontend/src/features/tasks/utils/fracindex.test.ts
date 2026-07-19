import { describe, expect, it } from 'vitest';
import { generateKeyBetween } from './fracindex';

describe('generateKeyBetween', () => {
  const cases: { name: string; left: string; right: string }[] = [
    { name: 'default', left: '', right: '' },
    { name: 'before first', left: '', right: '500000000000' },
    { name: 'after last', left: '500000000000', right: '' },
    { name: 'between adjacent', left: '500000000000', right: '500000000001' },
  ];

  it.each(cases)('$name', ({ left, right }) => {
    const got = generateKeyBetween(left, right);
    if (left !== '') expect(got > left).toBe(true);
    if (right !== '') expect(got < right).toBe(true);
  });

  it('handles repeated insertions between adjacent keys', () => {
    const left = '500000000000';
    let right = '500000000001';

    for (let i = 0; i < 32; i++) {
      const got = generateKeyBetween(left, right);
      expect(got > left && got < right).toBe(true);
      right = got;
    }
  });

  it('throws on invalid bounds', () => {
    expect(() => generateKeyBetween('b', 'a')).toThrow('invalid bounds');
    expect(() => generateKeyBetween('a', 'a')).toThrow('invalid bounds');
  });

  it.each([
    { left: '', right: '', want: 'O' },
    { left: '', right: '500000000000', want: '+' },
    { left: '500000000000', right: '', want: '50000000000W' },
    { left: '500000000000', right: '500000000001', want: '500000000000O' },
    { left: '!', right: '"', want: '!O' },
    { left: '', right: '!!', want: '!!O' },
    { left: '~', right: '', want: '~O' },
    { left: 'a', right: 'c', want: 'b' },
    { left: 'a', right: 'b', want: 'aO' },
  ])('matches Go output for ("$left", "$right")', ({ left, right, want }) => {
    expect(generateKeyBetween(left, right)).toBe(want);
  });
});

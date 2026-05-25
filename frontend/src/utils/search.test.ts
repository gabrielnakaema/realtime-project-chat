import { describe, expect, it } from 'vitest';
import { hasSearchQuery, normalizeSearchQuery } from './search';

describe('normalizeSearchQuery', () => {
  it('returns an empty string for missing values', () => {
    expect(normalizeSearchQuery()).toBe('');
    expect(normalizeSearchQuery(null)).toBe('');
  });

  it('trims leading and trailing whitespace', () => {
    expect(normalizeSearchQuery('  roadmap  ')).toBe('roadmap');
  });

  it('preserves meaningful spacing inside the query', () => {
    expect(normalizeSearchQuery('design system')).toBe('design system');
  });
});

describe('hasSearchQuery', () => {
  it('treats whitespace-only input as empty', () => {
    expect(hasSearchQuery('   ')).toBe(false);
    expect(hasSearchQuery('\n\t')).toBe(false);
  });

  it('detects usable search terms', () => {
    expect(hasSearchQuery('kanban')).toBe(true);
  });
});

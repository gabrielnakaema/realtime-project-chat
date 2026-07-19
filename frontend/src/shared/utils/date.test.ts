import { describe, expect, it } from 'vitest';
import {
  formatDateString,
  formatDateTime,
  formatLongDate,
  formatMonthDay,
  formatShortDate,
  formatTaskDueDate,
} from './date';

const sampleDate = '2026-06-02T15:45:00.000Z';

describe('date utils', () => {
  it('formats date-time values with the shared formatter', () => {
    expect(formatDateTime(sampleDate)).toBe(
      new Date(sampleDate).toLocaleDateString([], {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      }),
    );
  });

  it('formats short dates with a shared month-day-year shape', () => {
    expect(formatShortDate(sampleDate)).toBe(
      new Date(sampleDate).toLocaleDateString([], {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
      }),
    );
  });

  it('formats month-day strings for compact timeline and deadline UI', () => {
    expect(formatMonthDay(sampleDate)).toBe(
      new Date(sampleDate).toLocaleDateString([], {
        month: 'short',
        day: 'numeric',
      }),
    );
  });

  it('formats long dates for detail views', () => {
    expect(formatLongDate(sampleDate)).toBe(
      new Date(sampleDate).toLocaleDateString([], {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      }),
    );
  });

  it('supports custom fallback labels', () => {
    expect(formatDateTime(null, 'Never used')).toBe('Never used');
    expect(formatShortDate(undefined, 'No due date')).toBe('No due date');
  });

  it('keeps the legacy exports aligned with the shared helpers', () => {
    expect(formatDateString(sampleDate)).toBe(formatDateTime(sampleDate));
    expect(formatTaskDueDate(sampleDate)).toBe(formatShortDate(sampleDate));
  });
});

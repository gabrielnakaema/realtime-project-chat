const DATE_TIME_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
};

const SHORT_DATE_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
};

const MONTH_DAY_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  month: 'short',
  day: 'numeric',
};

const LONG_DATE_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
};

const formatDate = (date: string | null | undefined, options: Intl.DateTimeFormatOptions, fallback = '-'): string => {
  if (!date) {
    return fallback;
  }

  return new Date(date).toLocaleDateString([], options);
};

export const formatDateTime = (date: string | null | undefined, fallback = '-') => {
  return formatDate(date, DATE_TIME_FORMAT_OPTIONS, fallback);
};

export const formatShortDate = (date: string | null | undefined, fallback = '-') => {
  return formatDate(date, SHORT_DATE_FORMAT_OPTIONS, fallback);
};

export const formatMonthDay = (date: string | null | undefined, fallback = '-') => {
  return formatDate(date, MONTH_DAY_FORMAT_OPTIONS, fallback);
};

export const formatLongDate = (date: string | null | undefined, fallback = '-') => {
  return formatDate(date, LONG_DATE_FORMAT_OPTIONS, fallback);
};

export const formatDateString = (date: string | null): string => {
  return formatDateTime(date);
};

export const formatTaskDueDate = (dueDate: string | null) => {
  return formatShortDate(dueDate);
};

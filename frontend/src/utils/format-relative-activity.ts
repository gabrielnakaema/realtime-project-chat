import { differenceInCalendarDays } from 'date-fns';

export const formatRelativeActivityDateString = (isoString: string) => {
  const date = new Date(isoString);
  const now = new Date();
  const seconds = (now.getTime() - date.getTime()) / 1000;
  const days = differenceInCalendarDays(now, date);

  if (seconds <= 60) {
    return 'Just now';
  }

  if (days === 0) {
    return 'Today at ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  if (days === 1) {
    return 'Yesterday at ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  return date.toLocaleDateString([], {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

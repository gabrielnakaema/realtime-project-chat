export const formatRelativeActivityDateString = (isoString: string) => {
  const date = new Date(isoString);
  const now = new Date();
  const seconds = (now.getTime() - date.getTime()) / 1000;
  const days = Math.floor(seconds / (60 * 60 * 24));

  if (seconds <= 60) {
    return 'Just now';
  }

  if (days > 6) {
    return date.toLocaleDateString([], {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

  const divisions = [
    { amount: 60, name: 'second' },
    { amount: 60, name: 'minute' },
    { amount: 24, name: 'hour' },
    { amount: 7, name: 'day' },
  ];

  let duration = seconds;
  for (const division of divisions) {
    if (Math.abs(duration) < division.amount) {
      return rtf.format(Math.round(-duration), division.name as Intl.RelativeTimeFormatUnit);
    }
    duration /= division.amount;
  }

  return rtf.format(-Math.round(duration), 'day');
};

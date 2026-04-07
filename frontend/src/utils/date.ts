export const formatDateString = (date: string | null): string => {
  if (!date) return '-';
  return new Date(date).toLocaleDateString([], {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

export const formatTaskDueDate = (dueDate: string | null) => {
  if (!dueDate) {
    return '-';
  }
  return new Date(dueDate).toLocaleDateString([], {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
};

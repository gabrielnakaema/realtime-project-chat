export const priorityOptions = [
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
];

export const parseUniqueTags = (value?: string | null) =>
  Array.from(
    new Set(
      (value ?? '')
        .split(',')
        .map((tag) => tag.trim())
        .filter(Boolean),
    ),
  );

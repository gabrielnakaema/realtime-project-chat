export const DEFAULT_PROJECT_COLUMN_COLORS = [
  '#64748B',
  '#2563EB',
  '#059669',
  '#D97706',
  '#DC2626',
  '#0891B2',
] as const;

export const getDefaultProjectColumnColor = (index: number) =>
  DEFAULT_PROJECT_COLUMN_COLORS[index % DEFAULT_PROJECT_COLUMN_COLORS.length];

const normalizeHex = (hex: string) => {
  const value = hex.trim();
  return /^#[0-9A-Fa-f]{6}$/.test(value) ? value : '#64748B';
};

export const hexToRgba = (hex: string, alpha: number) => {
  const normalized = normalizeHex(hex).replace('#', '');
  const red = Number.parseInt(normalized.slice(0, 2), 16);
  const green = Number.parseInt(normalized.slice(2, 4), 16);
  const blue = Number.parseInt(normalized.slice(4, 6), 16);

  return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
};

export const buildProjectColumnSurface = (hex: string) => ({
  backgroundColor: hexToRgba(hex, 0.14),
  borderColor: hexToRgba(hex, 0.24),
  accentColor: normalizeHex(hex),
  badgeBackground: hexToRgba(hex, 0.18),
});

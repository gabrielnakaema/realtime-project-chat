export const normalizeSearchQuery = (query?: string | null) => {
  return query?.trim() ?? '';
};

export const hasSearchQuery = (query?: string | null) => {
  return normalizeSearchQuery(query).length > 0;
};

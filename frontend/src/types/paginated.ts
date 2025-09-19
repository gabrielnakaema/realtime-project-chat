export interface Paginated<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface CursorPaginated<T> {
  data: T[];
  hasNext: boolean;
}

export interface Paginated<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface CursorPaginated<T> {
  data: T[];
  has_next: boolean;
}

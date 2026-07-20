export interface Board {
  columns: Column;
}

export interface Column {
  id: string;
  project_id: string;
  title: string;
  description: string;
  color: string;
  columnId: string;
  isDoneColumn: boolean;
  total: number;
}

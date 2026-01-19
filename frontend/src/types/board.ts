import type { TaskStatus } from './task';

export interface Board {
  columns: Column;
}

export interface Column {
  id: string;
  project_id: string;
  title: string;
  color: string;
  status: TaskStatus;
  total: number;
}

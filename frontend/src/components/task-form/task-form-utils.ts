import { format, parseISO } from 'date-fns';
import type { Member } from '@/types/project';
import type { Task } from '@/types/task';
import type { ITaskForm } from '@/schemas/task-schema';

export const priorityOptions = [
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
];

export interface TaskMemberOption {
  label: string;
  value: string;
}

export const getTaskMemberOptions = (projectMembers: Member[]): TaskMemberOption[] =>
  projectMembers.map((member) => ({
    label: member.user.name,
    value: member.user.id,
  }));

export const formatTaskDependencyLabel = (task: Pick<Task, 'title' | 'code'>): string => {
  const code = task.code?.trim() || '';

  if (code) {
    return `${code} — ${task.title}`;
  }

  return task.title;
};

export const getTaskFormValues = (task: Task): ITaskForm => ({
  project_column_id: task.project_column_id,
  code: task.code || undefined,
  title: task.title,
  description: task.description,
  due_date: task.due_date ? format(parseISO(task.due_date), 'yyyy-MM-dd') : undefined,
  priority: task.priority,
  responsible_id: task.responsible_id,
  tags: task.tags?.join(',') || undefined,
  depends_on_task_ids: task.depends_on_task_ids ?? [],
});

export const parseUniqueTags = (value?: string | null) =>
  Array.from(
    new Set(
      (value ?? '')
        .split(',')
        .map((tag) => tag.trim())
        .filter(Boolean),
    ),
  );

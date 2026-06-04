import type { Member } from '@/types/project';

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

export const parseUniqueTags = (value?: string | null) =>
  Array.from(
    new Set(
      (value ?? '')
        .split(',')
        .map((tag) => tag.trim())
        .filter(Boolean),
    ),
  );

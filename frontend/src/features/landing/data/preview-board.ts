export type PreviewColumnId = 'backlog' | 'doing' | 'review' | 'done';
export type PreviewPriority = 'Low' | 'Medium' | 'High';

export interface PreviewTask {
  id: string;
  code: string;
  title: string;
  priority: PreviewPriority;
  owner: string;
  ownerName: string;
  tags: ReadonlyArray<string>;
  dueDate?: string;
  highlighted?: boolean;
}

export interface PreviewColumn {
  id: PreviewColumnId;
  name: string;
  color: string;
  isDone?: boolean;
}

export const previewColumns: ReadonlyArray<PreviewColumn> = [
  { id: 'backlog', name: 'Backlog', color: '#64748B' },
  { id: 'doing', name: 'Doing', color: '#2563EB' },
  { id: 'review', name: 'Review', color: '#D97706' },
  { id: 'done', name: 'Done', color: '#059669', isDone: true },
];

export type PreviewBoardState = Record<PreviewColumnId, ReadonlyArray<PreviewTask>>;

export const initialPreviewBoard: PreviewBoardState = {
  backlog: [
    {
      id: 'plat-101',
      code: 'PLAT-101',
      title: 'Map the first-run experience',
      priority: 'Medium',
      owner: 'AK',
      ownerName: 'Alex Kim',
      tags: ['research'],
    },
    {
      id: 'plat-105',
      code: 'PLAT-105',
      title: 'Draft the onboarding checklist',
      priority: 'Low',
      owner: 'MS',
      ownerName: 'Morgan Smith',
      tags: ['product'],
    },
    {
      id: 'plat-104',
      code: 'PLAT-104',
      title: 'Ship the new landing page',
      priority: 'Medium',
      owner: 'JS',
      ownerName: 'John S.',
      tags: ['frontend', 'launch'],
      dueDate: 'Jul 24',
      highlighted: true,
    },
  ],
  doing: [
    {
      id: 'plat-102',
      code: 'PLAT-102',
      title: 'Connect realtime activity',
      priority: 'High',
      owner: 'JL',
      ownerName: 'Jordan Lee',
      tags: ['backend', 'ws'],
      dueDate: 'Jul 22',
    },
    {
      id: 'plat-106',
      code: 'PLAT-106',
      title: 'Wire up presence indicators',
      priority: 'Medium',
      owner: 'AK',
      ownerName: 'Alex Kim',
      tags: ['frontend'],
    },
  ],
  review: [
    {
      id: 'plat-103',
      code: 'PLAT-103',
      title: 'Polish project navigation',
      priority: 'Low',
      owner: 'MS',
      ownerName: 'Morgan Smith',
      tags: ['frontend'],
    },
    {
      id: 'plat-107',
      code: 'PLAT-107',
      title: 'Audit the empty states',
      priority: 'Medium',
      owner: 'JL',
      ownerName: 'Jordan Lee',
      tags: ['design'],
    },
  ],
  done: [
    {
      id: 'plat-100',
      code: 'PLAT-100',
      title: 'Agree on launch scope',
      priority: 'High',
      owner: 'AK',
      ownerName: 'Alex Kim',
      tags: ['planning'],
    },
    {
      id: 'plat-099',
      code: 'PLAT-099',
      title: 'Set up the shared workspace',
      priority: 'Medium',
      owner: 'JS',
      ownerName: 'John S.',
      tags: ['setup'],
    },
  ],
};

export const previewMemberInitials = ['AK', 'JL', 'MS', 'JS'];

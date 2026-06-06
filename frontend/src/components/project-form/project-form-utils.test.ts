import { describe, expect, it } from 'vitest';
import { createDefaultProjectColumns, getProjectColumnKey, getProjectFormValues } from './project-form-utils';

describe('project form utils', () => {
  it('creates the default project columns with a single done column', () => {
    const columns = createDefaultProjectColumns();

    expect(columns).toHaveLength(3);
    expect(columns.map((column) => column.name)).toEqual(['Pending', 'Doing', 'Done']);
    expect(columns.map((column) => column.description)).toEqual(['', '', '']);
    expect(columns.filter((column) => column.is_done_column)).toHaveLength(1);
    expect(columns[2]?.is_done_column).toBe(true);
  });

  it('maps a project into form values', () => {
    const project = {
      id: 'project-1',
      user_id: 'user-1',
      name: 'Website redesign',
      description: '<p>Refresh the marketing site</p>',
      created_at: '2026-06-05T00:00:00Z',
      updated_at: '2026-06-05T00:00:00Z',
      members: [],
      columns: [
        {
          id: 'col-1',
          name: 'Backlog',
          description: 'Waiting for prioritization',
          color: '#123456',
          position: 0,
          is_done_column: false,
        },
        {
          id: 'col-2',
          name: 'Done',
          description: 'Finished work',
          color: '#654321',
          position: 1,
          is_done_column: true,
        },
      ],
    };

    expect(getProjectFormValues(project)).toEqual({
      name: 'Website redesign',
      description: '<p>Refresh the marketing site</p>',
      columns: [
        {
          id: 'col-1',
          name: 'Backlog',
          description: 'Waiting for prioritization',
          color: '#123456',
          is_done_column: false,
        },
        {
          id: 'col-2',
          name: 'Done',
          description: 'Finished work',
          color: '#654321',
          is_done_column: true,
        },
      ],
      deleted_columns: [],
    });
  });

  it('uses the column id when available and falls back to the index', () => {
    expect(getProjectColumnKey({ id: 'col-1' }, 0)).toBe('col-1');
    expect(getProjectColumnKey({}, 2)).toBe('new-column-2');
  });
});

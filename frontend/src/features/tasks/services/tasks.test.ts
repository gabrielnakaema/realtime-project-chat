import { beforeEach, describe, expect, it, vi } from 'vitest';
import { searchTasksForUser } from './tasks';
import { api } from '@/shared/services/api';

vi.mock('@/shared/services/api', () => ({
  api: {
    get: vi.fn(),
  },
}));

describe('searchTasksForUser', () => {
  beforeEach(() => {
    vi.mocked(api.get).mockReset();
  });

  it('sends the stable task id cursor without adding project filters', async () => {
    const page = { data: [], has_next: false, has_previous: false };
    vi.mocked(api.get).mockReturnValue({
      json: vi.fn().mockResolvedValue(page),
    } as never);

    await expect(
      searchTasksForUser({
        cursorDueDate: null,
        cursorTaskId: '0f6360c1-e6d2-480c-8174-986c01254072',
        cursorUpdatedAt: '2026-07-23T12:00:00.000Z',
        limit: 15,
        searchQuery: 'task',
      }),
    ).resolves.toEqual(page);

    expect(api.get).toHaveBeenCalledOnce();
    const [path, options] = vi.mocked(api.get).mock.calls[0];
    expect(path).toBe('tasks/search');
    expect(options?.searchParams?.toString()).toBe(
      'task_id=0f6360c1-e6d2-480c-8174-986c01254072&updated_at=2026-07-23T12%3A00%3A00.000Z&limit=15&query=task',
    );
  });
});

import { describe, expect, it } from 'vitest';
import { prependCommentToInfiniteData } from './use-task-comments';
import type { TaskComment } from '@/features/tasks/types/task';
import type { CursorPaginated } from '@/shared/types/paginated';
import type { InfiniteData } from '@tanstack/react-query';
import type { CommentsPageParam } from '@/features/tasks/hooks/task-comments-pagination';

type CommentsInfiniteData = InfiniteData<CursorPaginated<TaskComment>, CommentsPageParam>;

const makeComment = (overrides: Partial<TaskComment> & { id: string }): TaskComment => ({
  task: null,
  user: { id: 'user-1', name: 'Alice', email: 'alice@example.com' },
  content: '<p>Hello</p>',
  parent_comment_id: null,
  action_origin: null,
  created_at: '2026-06-04T10:00:00.000Z',
  updated_at: '2026-06-04T10:00:00.000Z',
  replies: [],
  ...overrides,
});

const makePage = (comments: TaskComment[]): CursorPaginated<TaskComment> => ({
  data: comments,
  has_next: false,
  has_previous: false,
});

const makeInfiniteData = (pages: CursorPaginated<TaskComment>[]): CommentsInfiniteData => ({
  pages,
  pageParams: pages.map(() => ({ direction: 'initial' }) as CommentsPageParam),
});

describe('prependCommentToInfiniteData', () => {
  it('prepends a new top-level comment to the first page', () => {
    const existing = makeComment({ id: 'c1' });
    const incoming = makeComment({ id: 'c2' });
    const current = makeInfiniteData([makePage([existing])]);

    const result = prependCommentToInfiniteData(current, incoming);

    expect(result.pages[0].data).toEqual([incoming, existing]);
  });

  it('does not add a duplicate top-level comment', () => {
    const comment = makeComment({ id: 'c1' });
    const current = makeInfiniteData([makePage([comment])]);

    const result = prependCommentToInfiniteData(current, comment);

    expect(result.pages[0].data).toHaveLength(1);
  });

  it('inserts a reply under the correct parent comment', () => {
    const parent = makeComment({ id: 'parent-1' });
    const reply = makeComment({ id: 'reply-1', parent_comment_id: 'parent-1' });
    const current = makeInfiniteData([makePage([parent])]);

    const result = prependCommentToInfiniteData(current, reply);

    expect(result.pages[0].data[0].replies).toEqual([reply]);
  });

  it('does not add a duplicate reply', () => {
    const reply = makeComment({ id: 'reply-1', parent_comment_id: 'parent-1' });
    const parent = makeComment({ id: 'parent-1', replies: [reply] });
    const current = makeInfiniteData([makePage([parent])]);

    const result = prependCommentToInfiniteData(current, reply);

    expect(result.pages[0].data[0].replies).toHaveLength(1);
  });

  it('inserts a reply into the correct nested parent', () => {
    const nestedReply = makeComment({ id: 'reply-2', parent_comment_id: 'reply-1' });
    const firstReply = makeComment({ id: 'reply-1', parent_comment_id: 'parent-1' });
    const parent = makeComment({ id: 'parent-1', replies: [firstReply] });
    const current = makeInfiniteData([makePage([parent])]);

    const result = prependCommentToInfiniteData(current, nestedReply);

    const insertedReply = result.pages[0].data[0].replies[0];
    expect(insertedReply.replies).toEqual([nestedReply]);
  });

  it('returns the same reference when the parent comment is not in any loaded page', () => {
    const comment = makeComment({ id: 'c1' });
    const replyToUnloadedParent = makeComment({ id: 'reply-1', parent_comment_id: 'not-loaded' });
    const current = makeInfiniteData([makePage([comment])]);

    const result = prependCommentToInfiniteData(current, replyToUnloadedParent);

    expect(result).toBe(current);
  });

  it('initialises an empty first page when current data is undefined', () => {
    const incoming = makeComment({ id: 'c1' });

    const result = prependCommentToInfiniteData(undefined, incoming);

    expect(result.pages[0].data).toEqual([incoming]);
  });
});

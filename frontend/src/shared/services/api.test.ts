import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

type AfterResponseHook = (request: Request, options: Record<string, unknown>, response: Response) => Promise<unknown>;

const mocks = vi.hoisted(() => ({
  afterResponseHook: undefined as unknown,
  attemptLogout: vi.fn(),
  attemptRefreshToken: vi.fn(),
  reload: vi.fn(),
}));

vi.mock('ky', () => {
  const ky = vi.fn();
  const create = vi.fn((options: { hooks: { afterResponse: unknown[] } }) => {
    mocks.afterResponseHook = options.hooks.afterResponse[0];
    return {};
  });

  return { default: Object.assign(ky, { create }) };
});

vi.mock('@/features/auth/services/auth', () => ({
  attemptLogout: mocks.attemptLogout,
  attemptRefreshToken: mocks.attemptRefreshToken,
}));

await import('./api');

describe('api refresh queue', () => {
  beforeEach(() => {
    mocks.attemptLogout.mockReset().mockResolvedValue(undefined);
    mocks.attemptRefreshToken.mockReset();
    mocks.reload.mockReset();
    vi.stubGlobal('window', { location: { reload: mocks.reload } });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('rejects every queued request when refreshing the token fails', async () => {
    let rejectRefresh!: (error: Error) => void;
    mocks.attemptRefreshToken.mockReturnValue(
      new Promise((_, reject) => {
        rejectRefresh = reject;
      }),
    );

    const afterResponse = mocks.afterResponseHook as AfterResponseHook;
    const unauthorizedResponse = new Response(null, { status: 401 });
    const firstRequest = afterResponse(new Request('http://localhost:3333/projects'), {}, unauthorizedResponse);
    const queuedRequest = afterResponse(
      new Request('http://localhost:3333/projects/activities'),
      {},
      unauthorizedResponse,
    );
    const refreshError = new Error('refresh failed');
    const firstRejection = expect(firstRequest).rejects.toBe(refreshError);
    const queuedRejection = expect(queuedRequest).rejects.toBe(refreshError);

    rejectRefresh(refreshError);

    await Promise.all([firstRejection, queuedRejection]);
    expect(mocks.attemptRefreshToken).toHaveBeenCalledOnce();
    expect(mocks.attemptLogout).toHaveBeenCalledOnce();
    expect(mocks.reload).toHaveBeenCalledOnce();
  });
});

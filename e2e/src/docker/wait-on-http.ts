export async function waitOnHttp(
  url: string,
  timeoutMs: number,
  intervalMs = 500
): Promise<void> {
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    try {
      const res = await fetch(url);
      if (res.ok) return;
    } catch {
      // connection refused / not listening yet — keep retrying
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  throw new Error(
    `Timed out after ${timeoutMs}ms waiting for ${url} to respond`
  );
}

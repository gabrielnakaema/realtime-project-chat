import crypto from 'node:crypto';
import type { APIRequestContext } from '@playwright/test';
import { expect } from '@playwright/test';

export interface TestUser {
  name: string;
  email: string;
  password: string;
}

export async function registerUser(
  request: APIRequestContext,
  backendURL: string,
  overrides: Partial<TestUser> = {}
): Promise<TestUser> {
  const user: TestUser = {
    name: overrides.name ?? 'E2E User',
    email: overrides.email ?? `e2e-${crypto.randomUUID()}@example.com`,
    password: overrides.password ?? 'Password123!',
  };

  const res = await request.post(`${backendURL}/users`, { data: user });
  expect(res.ok(), `failed to register test user: ${res.status()} ${await res.text()}`).toBeTruthy();

  return user;
}

import crypto from "node:crypto";
import { test as base, type Browser, type Page } from "@playwright/test";
import { registerUser, type TestUser } from "./test-user.js";

export const backendURL = `http://localhost:${
  process.env.E2E_BACKEND_PORT ?? "4333"
}`;

interface AuthFixtures {
  backendURL: string;
  testUser: TestUser;
  authenticatedPage: Page;
}

export async function loginAsUser(
  browser: Browser,
  user: TestUser
): Promise<Page> {
  const context = await browser.newContext();

  await context.request.post(`${backendURL}/auth/login`, {
    data: { email: user.email, password: user.password },
  });

  const page = await context.newPage();

  await page.goto("/projects");
  await page.getByText(`Welcome back, ${user.name}`).waitFor();

  return page;
}

export const test = base.extend<AuthFixtures>({
  backendURL: async ({}, use) => {
    await use(backendURL);
  },

  testUser: async ({ request, backendURL }, use, testInfo) => {
    const user = await registerUser(request, backendURL, {
      email: `e2e-w${testInfo.workerIndex}-${crypto.randomUUID()}@example.com`,
    });
    await use(user);
  },

  authenticatedPage: async ({ browser, testUser }, use) => {
    const page = await loginAsUser(browser, testUser);
    await use(page);
    await page.context().close();
  },
});

export { expect } from "@playwright/test";

import crypto from "node:crypto";
import {
  expect,
  test as base,
  type Browser,
  type BrowserContext,
  type Page,
} from "@playwright/test";
import { registerUser, type TestUser } from "./test-user.js";

const gatewayPort = process.env.E2E_RESOLVED_GATEWAY_PORT;

if (!gatewayPort) {
  throw new Error(
    "E2E_RESOLVED_GATEWAY_PORT was not set by the global E2E setup"
  );
}

export const backendURL = `http://localhost:${gatewayPort}`;

interface AuthFixtures {
  backendURL: string;
  testUser: TestUser;
  authenticatedPage: Page;
}

const toastClickThroughScript = `
  const styleId = "e2e-toast-click-through";

  const installStyle = () => {
    if (document.getElementById(styleId)) {
      return;
    }

    const style = document.createElement("style");
    style.id = styleId;
    style.textContent = ".Toastify, .Toastify * { pointer-events: none !important; }";
    document.head.appendChild(style);
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", installStyle, {
      once: true,
    });
  } else {
    installStyle();
  }
`;

async function makeToastsClickThrough(target: BrowserContext | Page) {
  await target.addInitScript(toastClickThroughScript);
}

export async function expectToast(page: Page, message: string | RegExp) {
  await expect(page.getByText(message)).toBeVisible();
}

export async function openProjectMembersSettings(page: Page) {
  await page.getByRole("link", { name: "Settings" }).click();
  await page.getByRole("link", { name: "Members", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Settings", exact: true })
  ).toBeVisible();
}

export async function addProjectMember(page: Page, email: string) {
  await openProjectMembersSettings(page);
  await page.getByLabel("Email address").fill(email);
  await page.getByRole("button", { name: "Add member" }).click();
  await expectToast(page, "Member added successfully");
  await page.getByRole("link", { name: "Back to board" }).click();
}

export async function loginAsUser(
  browser: Browser,
  user: TestUser
): Promise<Page> {
  const context = await browser.newContext();
  await makeToastsClickThrough(context);

  await context.request.post(`${backendURL}/auth/login`, {
    data: { email: user.email, password: user.password },
  });

  const page = await context.newPage();

  await page.goto("/projects");
  await page.getByRole("heading", { name: "Projects", exact: true }).waitFor();

  return page;
}

export const test = base.extend<AuthFixtures>({
  page: async ({ page }, use) => {
    await makeToastsClickThrough(page);
    await use(page);
  },

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

export { expect };

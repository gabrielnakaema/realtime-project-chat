import crypto from "node:crypto";
import {
  expect,
  test as base,
  type Browser,
  type BrowserContext,
  type Locator,
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

export interface ProjectRepositorySettings {
  url?: string;
  owner?: string;
  name?: string;
  defaultBranch?: string;
  branchNamePrefix?: string;
}

export interface ProjectColumnSettings {
  name: string;
  description?: string;
  color?: string;
  isDone?: boolean;
}

export interface CreateProjectThroughUIOptions {
  description?: string;
  repository?: ProjectRepositorySettings;
  columns?: ProjectColumnSettings[];
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
  const toast = page.getByRole("alert").filter({ hasText: message }).last();
  await expect(toast).toBeVisible();
}

function projectColumnEditor(container: Page | Locator, index: number) {
  return container
    .locator(`#column-${index}`)
    .locator(`xpath=ancestor::*[.//*[@id="column-description-${index}"]][1]`);
}

export async function createProjectThroughUI(
  page: Page,
  name: string,
  options: CreateProjectThroughUIOptions = {}
) {
  const {
    description = "Created by an end-to-end test.",
    repository,
    columns,
  } = options;

  if (columns?.length === 0) {
    throw new Error("A project needs at least one column");
  }

  if (columns && columns.filter((column) => column.isDone).length > 1) {
    throw new Error("A project can only have one done column");
  }

  await page.getByRole("button", { name: "New project" }).first().click();

  const dialog = page.getByRole("dialog", { name: "Create project" });
  await dialog.locator("#name").fill(name);
  await dialog.locator("#description").pressSequentially(description);

  const repositoryFields = {
    repository_url: repository?.url,
    repository_owner: repository?.owner,
    repository_name: repository?.name,
    default_branch: repository?.defaultBranch,
    branch_name_prefix: repository?.branchNamePrefix,
  };

  for (const [field, value] of Object.entries(repositoryFields)) {
    if (value !== undefined) {
      await dialog.locator(`#${field}`).fill(value);
    }
  }

  if (columns) {
    let columnCount = 3;

    while (columnCount < columns.length) {
      await dialog.getByRole("button", { name: "Add column" }).click();
      columnCount += 1;
    }

    while (columnCount > columns.length) {
      const index = columnCount - 1;
      await projectColumnEditor(dialog, index)
        .getByRole("button", { name: /^Delete / })
        .click();
      columnCount -= 1;
    }

    for (const [index, column] of columns.entries()) {
      const editor = projectColumnEditor(dialog, index);
      await editor.locator(`#column-${index}`).fill(column.name);

      if (column.description !== undefined) {
        await editor
          .locator(`#column-description-${index}`)
          .fill(column.description);
      }

      if (column.color !== undefined) {
        await editor.locator(`#column-color-${index}`).fill(column.color);
      }

      if (column.isDone) {
        await editor
          .getByRole("button", { name: /^(Mark as done|Done column)$/ })
          .click();
      }
    }
  }

  await dialog.getByRole("button", { name: "Create project" }).click();
  await expectToast(page, "Project created successfully");
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

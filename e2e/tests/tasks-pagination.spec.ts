import crypto from "node:crypto";
import {
  createProjectThroughUI,
  expect,
  test,
} from "../src/fixtures/authenticated-page.js";
import type { Page } from "@playwright/test";

async function createProject(page: Page, name: string) {
  await createProjectThroughUI(page, name);
  await page.getByRole("link", { name: new RegExp(name) }).click();
  await expect(page.getByRole("button", { name, exact: true })).toBeVisible();
}

async function apiAuthHeaders(
  page: Page,
  backendURL: string
): Promise<Record<string, string>> {
  const res = await page.request.post(`${backendURL}/auth/refresh-token`);
  expect(
    res.ok(),
    `failed to refresh access token: ${res.status()}`
  ).toBeTruthy();
  const { access_token: accessToken } = (await res.json()) as {
    access_token: string;
  };
  return { Authorization: `Bearer ${accessToken}` };
}

interface ProjectContext {
  projectId: string;
  columnId: (name: string) => string;
}

async function getProjectContext(
  page: Page,
  backendURL: string,
  headers: Record<string, string>
): Promise<ProjectContext> {
  const match = new URL(page.url()).pathname.match(
    /\/projects\/([0-9a-fA-F-]{36})/
  );
  if (!match) {
    throw new Error(`could not resolve project id from url: ${page.url()}`);
  }
  const projectId = match[1];

  const res = await page.request.get(`${backendURL}/projects/${projectId}`, {
    headers,
  });
  expect(
    res.ok(),
    `failed to load project ${projectId}: ${res.status()}`
  ).toBeTruthy();
  const project = (await res.json()) as {
    columns: { id: string; name: string }[];
  };

  return {
    projectId,
    columnId: (name: string) => {
      const column = project.columns.find((c) => c.name === name);
      if (!column) {
        throw new Error(`column "${name}" not found on project ${projectId}`);
      }
      return column.id;
    },
  };
}

async function seedTasksInColumn(
  page: Page,
  backendURL: string,
  projectId: string,
  columnId: string,
  titles: string[],
  headers: Record<string, string>
) {
  for (let i = titles.length - 1; i >= 0; i--) {
    const res = await page.request.post(`${backendURL}/tasks`, {
      headers,
      data: {
        project_id: projectId,
        project_column_id: columnId,
        title: titles[i],
        description: "Seeded by the task pagination e2e test",
        priority: "medium",
      },
    });
    expect(
      res.ok(),
      `failed to seed task "${titles[i]}": ${res.status()} ${await res.text()}`
    ).toBeTruthy();
  }
}

function boardColumn(page: Page, columnName: string) {
  return page
    .getByRole("button", { name: `Open actions for ${columnName}` })
    .locator(
      "xpath=ancestor::div[.//div[contains(@class, 'overflow-y-auto')]][1]"
    );
}

function columnTaskHeadings(page: Page, columnName: string) {
  return boardColumn(page, columnName).getByRole("heading", { level: 4 });
}

function columnCountBadge(page: Page, columnName: string) {
  return boardColumn(page, columnName).locator("span.inline-flex").first();
}

function taskCardInColumn(page: Page, columnName: string, title: string) {
  return boardColumn(page, columnName).getByText(title, { exact: true });
}

async function loadAllTasksInColumn(
  page: Page,
  columnName: string,
  expectedCount: number
) {
  const scroller = boardColumn(page, columnName).locator(".overflow-y-auto");

  await expect
    .poll(
      async () => {
        await scroller.evaluate((el) => el.scrollTo({ top: el.scrollHeight }));
        return columnTaskHeadings(page, columnName).count();
      },
      { timeout: 20000, intervals: [150, 300, 500, 800] }
    )
    .toBe(expectedCount);
}

async function dragTaskToColumn(
  page: Page,
  taskTitle: string,
  columnName: string
) {
  const draggableTask = taskCardInColumn(page, columnName, taskTitle)
    .or(page.getByText(taskTitle, { exact: true }))
    .first()
    .locator("xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]");
  const targetColumn = boardColumn(page, columnName);

  const moveResponsePromise = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      /\/tasks\/[^/]+\/move$/.test(new URL(response.url()).pathname)
  );

  await draggableTask.dragTo(targetColumn, {
    targetPosition: { x: 50, y: 200 },
  });

  const moveResponse = await moveResponsePromise;
  expect(moveResponse.ok()).toBe(true);
}

async function dragTaskBeforeTask(
  page: Page,
  taskTitle: string,
  targetTaskTitle: string
) {
  const draggableTask = page
    .getByText(taskTitle, { exact: true })
    .locator("xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]");
  const targetTask = page
    .getByText(targetTaskTitle, { exact: true })
    .locator("xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]");

  const moveResponsePromise = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      /\/tasks\/[^/]+\/move$/.test(new URL(response.url()).pathname)
  );

  await draggableTask.dragTo(targetTask, { targetPosition: { x: 40, y: 1 } });

  const moveResponse = await moveResponsePromise;
  expect(moveResponse.ok()).toBe(true);
}

test.describe("task board pagination", () => {
  test("loads a long column in pages as the user scrolls, while the count badge shows the full total", async ({
    authenticatedPage: page,
    backendURL,
  }) => {
    test.slow();
    const projectName = `E2E Pagination Project ${crypto.randomUUID()}`;
    const runId = crypto.randomUUID().slice(0, 8);
    const total = 22;
    const titles = Array.from(
      { length: total },
      (_, k) => `E2E Bulk ${runId} ${String(k + 1).padStart(3, "0")}`
    );

    await createProject(page, projectName);
    const headers = await apiAuthHeaders(page, backendURL);
    const { projectId, columnId } = await getProjectContext(
      page,
      backendURL,
      headers
    );
    await seedTasksInColumn(
      page,
      backendURL,
      projectId,
      columnId("Pending"),
      titles,
      headers
    );
    await page.reload();

    await expect(columnCountBadge(page, "Pending")).toHaveText(String(total));

    const headings = columnTaskHeadings(page, "Pending");
    expect(await headings.count()).toBeLessThan(total);
    await expect(
      taskCardInColumn(page, "Pending", titles[total - 1])
    ).toHaveCount(0);

    await loadAllTasksInColumn(page, "Pending", total);
    await expect(
      taskCardInColumn(page, "Pending", titles[total - 1])
    ).toBeVisible();
    await expect(headings).toHaveText(titles);
  });

  test("moving tasks in a fully paginated column stays consistent and persists across reload", async ({
    authenticatedPage: page,
    backendURL,
  }) => {
    test.slow();
    const projectName = `E2E Pagination Move Project ${crypto.randomUUID()}`;
    const runId = crypto.randomUUID().slice(0, 8);
    const total = 18;
    const titles = Array.from(
      { length: total },
      (_, k) => `E2E Move ${runId} ${String(k + 1).padStart(3, "0")}`
    );

    await createProject(page, projectName);
    const headers = await apiAuthHeaders(page, backendURL);
    const { projectId, columnId } = await getProjectContext(
      page,
      backendURL,
      headers
    );
    await seedTasksInColumn(
      page,
      backendURL,
      projectId,
      columnId("Pending"),
      titles,
      headers
    );
    await page.reload();

    await expect(columnCountBadge(page, "Pending")).toHaveText(String(total));
    await loadAllTasksInColumn(page, "Pending", total);

    const movedTitle = titles[total - 1];
    await dragTaskToColumn(page, movedTitle, "Done");

    await expect(taskCardInColumn(page, "Pending", movedTitle)).toHaveCount(0);
    await expect(taskCardInColumn(page, "Done", movedTitle)).toBeVisible();
    await expect(columnCountBadge(page, "Pending")).toHaveText(
      String(total - 1)
    );
    await expect(columnCountBadge(page, "Done")).toHaveText("1");

    const lastTitle = titles[total - 2];
    const penultimateTitle = titles[total - 3];
    await dragTaskBeforeTask(page, lastTitle, penultimateTitle);

    const pendingCount = total - 1;
    const pendingHeadings = columnTaskHeadings(page, "Pending");

    await expect(pendingHeadings.nth(pendingCount - 2)).toHaveText(lastTitle);
    await expect(pendingHeadings.nth(pendingCount - 1)).toHaveText(
      penultimateTitle
    );

    await page.reload();
    await expect(columnCountBadge(page, "Pending")).toHaveText(
      String(pendingCount)
    );
    await expect(columnCountBadge(page, "Done")).toHaveText("1");
    await expect(taskCardInColumn(page, "Done", movedTitle)).toBeVisible();

    await loadAllTasksInColumn(page, "Pending", pendingCount);
    await expect(pendingHeadings.nth(pendingCount - 2)).toHaveText(lastTitle);
    await expect(pendingHeadings.nth(pendingCount - 1)).toHaveText(
      penultimateTitle
    );
    await expect(taskCardInColumn(page, "Pending", movedTitle)).toHaveCount(0);
  });
});

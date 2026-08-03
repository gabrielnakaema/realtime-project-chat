import crypto from "node:crypto";
import {
  createProjectThroughUI,
  expect,
  expectToast,
  test,
} from "../src/fixtures/authenticated-page.js";

test("project owner sees task creation in recent activity and can return to the project", async ({
  authenticatedPage: page,
  testUser,
}) => {
  const projectName = `E2E Activity Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Activity Task ${crypto.randomUUID()}`;

  await createProjectThroughUI(page, projectName, {
    description: "Created by the project activity e2e test",
  });

  await page.getByRole("link", { name: projectName }).click();
  await page.getByRole("button", { name: "Open actions for Doing" }).click();
  await page.getByRole("menuitem", { name: "Add task" }).click();

  const createTaskDialog = page.getByRole("dialog", {
    name: "Create task",
  });
  await createTaskDialog.locator("#title").fill(taskTitle);
  await createTaskDialog
    .locator("#description")
    .pressSequentially("A task created to verify the activity feed.");
  await createTaskDialog.locator("#priority").click();
  await page.getByRole("option", { name: "Medium", exact: true }).click();
  await createTaskDialog.getByRole("button", { name: "Create task" }).click();
  await expectToast(page, "Task created successfully");

  await page.goto("/projects");
  const activityFeed = page.getByRole("region", { name: "Recent Activity" });
  await expect(activityFeed.getByText(taskTitle, { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await expect(
    activityFeed.getByText(`${testUser.name} created task`, { exact: false })
  ).toBeVisible();

  const taskActivity = activityFeed.getByRole("article").filter({
    hasText: taskTitle,
  });
  await taskActivity
    .getByRole("link", { name: projectName, exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: projectName, exact: true })
  ).toBeVisible();
});

test("project owner sees a task move in recent activity and can return to the project", async ({
  authenticatedPage: page,
  testUser,
}) => {
  const projectName = `E2E Updated Activity Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Updated Activity Task ${crypto.randomUUID()}`;

  await createProjectThroughUI(page, projectName, {
    description: "Created by the task update activity e2e test.",
  });

  await page.getByRole("link", { name: projectName }).click();
  await page.getByRole("button", { name: "Open actions for Doing" }).click();
  await page.getByRole("menuitem", { name: "Add task" }).click();
  const createTaskDialog = page.getByRole("dialog", {
    name: "Create task",
  });
  await createTaskDialog.locator("#title").fill(taskTitle);
  await createTaskDialog
    .locator("#description")
    .pressSequentially("A task moved to verify the updated activity feed.");
  await createTaskDialog.locator("#priority").click();
  await page.getByRole("option", { name: "High", exact: true }).click();
  await createTaskDialog.getByRole("button", { name: "Create task" }).click();
  await expectToast(page, "Task created successfully");

  const draggableTask = page
    .getByText(taskTitle, { exact: true })
    .locator("xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]");
  const doneColumn = page
    .getByRole("button", { name: "Open actions for Done" })
    .locator(
      "xpath=ancestor::div[.//div[contains(@class, 'overflow-y-auto')]][1]"
    );
  const moveResponsePromise = page.waitForResponse((response) => {
    const url = new URL(response.url());

    return (
      response.request().method() === "PATCH" &&
      /\/tasks\/[^/]+\/move$/.test(url.pathname)
    );
  });
  await draggableTask.dragTo(doneColumn, {
    targetPosition: { x: 50, y: 200 },
  });
  expect((await moveResponsePromise).ok()).toBe(true);

  await page.goto("/projects");
  const activityFeed = page.getByRole("region", { name: "Recent Activity" });
  const updatedTaskActivity = activityFeed
    .getByRole("article")
    .filter({
      hasText: taskTitle,
    })
    .filter({ hasText: `${testUser.name} updated task` });
  await expect(updatedTaskActivity).toBeVisible({ timeout: 15_000 });
  await expect(
    updatedTaskActivity.getByText(taskTitle, { exact: true })
  ).toBeVisible();

  await updatedTaskActivity
    .getByRole("link", { name: projectName, exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: projectName, exact: true })
  ).toBeVisible();
});

import crypto from "node:crypto";
import {
  expect,
  expectToast,
  loginAsUser,
  test,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";
import type { Page } from "@playwright/test";

function boardColumn(page: Page, columnName: string) {
  return page
    .getByRole("button", { name: `Open actions for ${columnName}` })
    .locator("xpath=ancestor::div[contains(@class, 'min-w-84')][1]");
}

test("project collaborators see task creation and movement on their board in real time", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "Realtime Collaborator",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Realtime Task Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Realtime Task ${crypto.randomUUID()}`;

  await ownerPage
    .getByRole("banner")
    .getByRole("button", { name: "Create project" })
    .click();
  const createProjectDialog = ownerPage.getByRole("dialog", {
    name: "Create project",
  });
  await createProjectDialog.locator("#name").fill(projectName);
  await createProjectDialog
    .locator("#description")
    .pressSequentially("Created by the real-time task board e2e test.");
  await createProjectDialog
    .getByRole("button", { name: "Create project" })
    .click();
  await expectToast(ownerPage, "Project created successfully");

  await ownerPage.getByRole("link", { name: projectName }).click();
  const projectId = ownerPage.url().split("/projects/")[1];
  await ownerPage.getByTitle("Add project member").click();
  const addMemberDialog = ownerPage.getByRole("dialog", {
    name: "Add project member",
  });
  await addMemberDialog.getByLabel("Email").fill(member.email);
  await addMemberDialog.getByRole("button", { name: "Add member" }).click();
  await expectToast(ownerPage, "Member added successfully");

  await ownerPage.reload();
  await memberPage.goto(`/projects/${projectId}`);
  await expect(
    memberPage.getByRole("heading", { name: projectName, exact: true })
  ).toBeVisible();

  await expect(
    ownerPage.locator("div.border-green-500").filter({ hasText: "R" })
  ).toBeVisible({ timeout: 15_000 });

  await ownerPage
    .getByRole("button", { name: "Open actions for Doing" })
    .click();
  await ownerPage.getByRole("menuitem", { name: "Add task" }).click();
  const createTaskDialog = ownerPage.getByRole("dialog", {
    name: "Create task",
  });
  await createTaskDialog.locator("#title").fill(taskTitle);
  await createTaskDialog
    .locator("#description")
    .pressSequentially(
      "A task created for a collaborator who is already viewing the board."
    );
  await createTaskDialog.locator("#priority").click();
  await ownerPage.getByRole("option", { name: "High", exact: true }).click();
  await createTaskDialog.getByRole("button", { name: "Create task" }).click();
  await expectToast(ownerPage, "Task created successfully");

  await expect(
    boardColumn(memberPage, "Doing").getByText(taskTitle, { exact: true })
  ).toBeVisible({ timeout: 15_000 });

  const draggableTask = ownerPage
    .getByText(taskTitle, { exact: true })
    .locator("xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]");
  const doneColumn = boardColumn(ownerPage, "Done");
  const moveResponsePromise = ownerPage.waitForResponse((response) => {
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

  await expect(
    boardColumn(memberPage, "Doing").getByText(taskTitle, { exact: true })
  ).toHaveCount(0);
  await expect(
    boardColumn(memberPage, "Done").getByText(taskTitle, { exact: true })
  ).toBeVisible({ timeout: 15_000 });

  await memberPage.context().close();
});

import crypto from "node:crypto";
import type { Page } from "@playwright/test";
import {
  addProjectMember,
  expect,
  expectToast,
  loginAsUser,
  openProjectMembersSettings,
  test,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";

async function createProject(page: Page, projectName: string) {
  await page.getByRole("button", { name: "New project" }).first().click();
  const dialog = page.getByRole("dialog", { name: "Create project" });
  await dialog.locator("#name").fill(projectName);
  await dialog
    .locator("#description")
    .pressSequentially("Created by a project realtime e2e test.");
  await dialog.getByRole("button", { name: "Create project" }).click();
  await expectToast(page, "Project created successfully");
  await page.getByRole("link", { name: projectName }).click();

  return page.url().split("/projects/")[1];
}

async function addMember(ownerPage: Page, email: string) {
  await addProjectMember(ownerPage, email);
}

function projectListLink(page: Page, projectName: string) {
  return page
    .getByRole("region", { name: "Projects" })
    .getByRole("link", { name: new RegExp(projectName) });
}

async function createTask(page: Page, taskTitle: string) {
  await page.getByRole("button", { name: "Open actions for Doing" }).click();
  await page.getByRole("menuitem", { name: "Add task" }).click();
  const dialog = page.getByRole("dialog", { name: "Create task" });
  await dialog.locator("#title").fill(taskTitle);
  await dialog
    .locator("#description")
    .pressSequentially("Realtime access check.");
  await dialog.locator("#priority").click();
  await page.getByRole("option", { name: "Medium", exact: true }).click();

  const responsePromise = page.waitForResponse((response) => {
    const url = new URL(response.url());
    return response.request().method() === "POST" && url.pathname === "/tasks";
  });
  await dialog.getByRole("button", { name: "Create task" }).click();
  const response = await responsePromise;
  expect(response.ok()).toBe(true);
  await expectToast(page, "Task created successfully");

  return (await response.json()) as { id: string };
}

async function removeMember(ownerPage: Page, email: string) {
  await openProjectMembersSettings(ownerPage);
  const memberRow = ownerPage.locator("article").filter({ hasText: email });
  await memberRow
    .getByRole("button", { name: /Remove .+ from project/ })
    .click();
  const removeDialog = ownerPage.getByRole("dialog", {
    name: "Remove member from project?",
  });
  await removeDialog.getByRole("button", { name: "Remove member" }).click();
  await expect(removeDialog).toHaveCount(0);
}

test("project metadata and workflow changes reach an open collaborator board without reload", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "Realtime Workflow Collaborator",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Realtime Workflow ${crypto.randomUUID()}`;
  const renamedProject = `E2E Renamed Workflow ${crypto.randomUUID()}`;
  const renamedColumn = "Ready for work";
  const projectId = await createProject(ownerPage, projectName);

  await addMember(ownerPage, member.email);
  await memberPage.goto(`/projects/${projectId}`);
  await expect(
    memberPage.getByRole("button", { name: projectName, exact: true })
  ).toBeVisible();
  await expect(
    ownerPage.locator("div.border-success").filter({ hasText: "R" })
  ).toBeVisible({ timeout: 15_000 });

  await ownerPage.getByRole("link", { name: "Settings" }).click();
  const generalSettings = ownerPage.locator("#general-project-settings");
  await generalSettings.locator("#name").fill(renamedProject);
  await generalSettings.getByRole("button", { name: "Save changes" }).click();
  await expectToast(ownerPage, "Project saved successfully");

  await ownerPage.getByRole("link", { name: "Columns", exact: true }).click();
  const columnSettings = ownerPage.locator("#columns-project-settings");
  await columnSettings.locator("#column-name-0").fill(renamedColumn);
  await columnSettings
    .locator("#column-description-0")
    .fill("Updated while another member watches the board.");
  await columnSettings.getByRole("button", { name: "Move Done up" }).click();
  await expect(
    columnSettings.locator('[data-slot="accordion-trigger"]')
  ).toHaveText([/Ready for work/, /Done/, /Doing/]);
  await columnSettings.getByRole("button", { name: "Save changes" }).click();
  await expectToast(ownerPage, "Project columns saved successfully");

  await expect(
    memberPage.getByRole("button", { name: renamedProject, exact: true })
  ).toBeVisible({ timeout: 15_000 });
  await expect(memberPage.getByRole("heading", { level: 3 })).toHaveText([
    renamedColumn,
    "Done",
    "Doing",
  ]);
  await expect(
    memberPage.getByRole("button", { name: projectName })
  ).toHaveCount(0);

  await memberPage.context().close();
});

test("member addition propagates to project, chat, notification, members, and websocket consumers without reload", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
  testUser,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "Realtime Added Member",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Realtime Addition ${crypto.randomUUID()}`;
  const taskTitle = `E2E Post-add Task ${crypto.randomUUID()}`;
  await createProject(ownerPage, projectName);

  await expect(projectListLink(memberPage, projectName)).toHaveCount(0);
  await addMember(ownerPage, member.email);

  await expect(projectListLink(memberPage, projectName)).toBeVisible({
    timeout: 15_000,
  });
  await memberPage.getByRole("button", { name: "Notifications" }).click();
  await expect(
    memberPage
      .getByRole("button")
      .filter({ hasText: `${testUser.name} added you to ${projectName}.` })
  ).toBeVisible({ timeout: 15_000 });
  await memberPage.keyboard.press("Escape");

  await projectListLink(memberPage, projectName).click();
  await openProjectMembersSettings(memberPage);
  await expect(memberPage.getByText("MEMBERS • 2")).toBeVisible();
  await expect(memberPage.getByText(member.email)).toBeVisible();
  await memberPage.getByRole("link", { name: "Back to board" }).click();

  await memberPage.getByRole("link", { name: "Chat", exact: true }).click();
  await expect(memberPage.getByPlaceholder("Message...")).toBeEnabled({
    timeout: 15_000,
  });
  await memberPage.getByRole("link", { name: "Go back" }).click();
  await expect(
    ownerPage.locator("div.border-success").filter({ hasText: "R" })
  ).toBeVisible({ timeout: 15_000 });

  await createTask(ownerPage, taskTitle);
  await expect(memberPage.getByText(taskTitle, { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await expect(
    ownerPage.locator("div.border-success").filter({ hasText: "R" })
  ).toBeVisible({ timeout: 15_000 });

  await memberPage.context().close();
});

test("member removal revokes open UI, chat, task, and websocket access without reload", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "Realtime Removed Member",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Realtime Removal ${crypto.randomUUID()}`;
  const taskTitle = `E2E Pre-removal Task ${crypto.randomUUID()}`;
  const projectId = await createProject(ownerPage, projectName);
  await addMember(ownerPage, member.email);
  await expect(projectListLink(memberPage, projectName)).toBeVisible({
    timeout: 15_000,
  });
  await projectListLink(memberPage, projectName).click();

  const task = await createTask(ownerPage, taskTitle);
  await expect(memberPage.getByText(taskTitle, { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await memberPage.getByRole("link", { name: "Chat", exact: true }).click();
  await expect(memberPage.getByPlaceholder("Message...")).toBeEnabled({
    timeout: 15_000,
  });
  await memberPage.getByRole("link", { name: "Go back" }).click();
  await expect(
    ownerPage.locator("div.border-success").filter({ hasText: "R" })
  ).toBeVisible({ timeout: 15_000 });

  await removeMember(ownerPage, member.email);

  await expect(memberPage).toHaveURL(/\/projects\/?$/, { timeout: 15_000 });
  await expect(projectListLink(memberPage, projectName)).toHaveCount(0);
  await expect(
    memberPage.getByRole("button", { name: projectName, exact: true })
  ).toHaveCount(0);

  const chatResponsePromise = memberPage.waitForResponse(
    (response) =>
      response.url().startsWith(backendURL) &&
      new URL(response.url()).pathname === `/projects/${projectId}/chat`
  );
  await memberPage.goto(`/projects/${projectId}/chat`);
  expect((await chatResponsePromise).status()).toBe(403);

  const taskResponsePromise = memberPage.waitForResponse(
    (response) =>
      response.url().startsWith(backendURL) &&
      new URL(response.url()).pathname === `/tasks/${task.id}`
  );
  const projectResponsePromise = memberPage.waitForResponse(
    (response) =>
      response.url().startsWith(backendURL) &&
      new URL(response.url()).pathname === `/projects/${projectId}`
  );
  await memberPage.goto(`/projects/${projectId}?taskId=${task.id}`);
  expect((await taskResponsePromise).status()).toBe(403);
  expect((await projectResponsePromise).status()).toBe(403);

  await memberPage.context().close();
});

test("project deletion redirects an open collaborator and revokes access without reload", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "Realtime Deletion Collaborator",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Realtime Deletion ${crypto.randomUUID()}`;
  const projectId = await createProject(ownerPage, projectName);
  await addMember(ownerPage, member.email);

  await expect(projectListLink(memberPage, projectName)).toBeVisible({
    timeout: 15_000,
  });
  await projectListLink(memberPage, projectName).click();
  await memberPage.getByRole("link", { name: "Settings" }).click();
  await expect(
    memberPage.locator("#general-project-settings")
  ).toBeVisible();

  await ownerPage.getByRole("link", { name: "Settings" }).click();
  await ownerPage.getByRole("button", { name: "Delete" }).click();
  const confirmDialog = ownerPage.getByRole("dialog", {
    name: `Delete ${projectName}?`,
  });
  await confirmDialog.getByRole("button", { name: "Delete project" }).click();
  await expectToast(ownerPage, "Project deleted successfully");

  await expect(memberPage).toHaveURL(/\/projects\/?$/, { timeout: 15_000 });
  await expectToast(memberPage, "This project was deleted.");
  await expect(projectListLink(memberPage, projectName)).toHaveCount(0);

  const projectResponsePromise = memberPage.waitForResponse(
    (response) =>
      response.url().startsWith(backendURL) &&
      new URL(response.url()).pathname === `/projects/${projectId}`
  );
  await memberPage.goto(`/projects/${projectId}`);
  expect((await projectResponsePromise).status()).toBe(404);

  await memberPage.context().close();
});

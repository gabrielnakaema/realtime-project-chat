import crypto from "node:crypto";
import type { Locator, Page } from "@playwright/test";
import {
  expect,
  expectToast,
  loginAsUser,
  test,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";

async function createProject(page: Page, projectName: string) {
  await page
    .getByRole("banner")
    .getByRole("button", { name: "Create project" })
    .click();
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
  await ownerPage.getByTitle("Add project member").click();
  const dialog = ownerPage.getByRole("dialog", { name: "Add project member" });
  await dialog.getByLabel("Email").fill(email);
  await dialog.getByRole("button", { name: "Add member" }).click();
  await expectToast(ownerPage, "Member added successfully");
}

function projectListLink(page: Page, projectName: string) {
  return page
    .getByRole("heading", { name: "Your Projects" })
    .locator("..")
    .getByRole("link", { name: new RegExp(projectName) });
}

function columnEditor(dialog: Locator, index: number) {
  return dialog
    .locator(`#column-${index}`)
    .locator(`xpath=ancestor::*[.//*[@id="column-description-${index}"]][1]`);
}

async function createTask(page: Page, taskTitle: string) {
  await page.getByRole("button", { name: "Open actions for Doing" }).click();
  await page.getByRole("menuitem", { name: "Add task" }).click();
  const dialog = page.getByRole("dialog", { name: "Create task" });
  await dialog.locator("#title").fill(taskTitle);
  await dialog.locator("#description").pressSequentially("Realtime access check.");
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
  await ownerPage.getByTitle("View project members").click();
  const membersDialog = ownerPage.getByRole("dialog", { name: "Project members" });
  const memberRow = membersDialog.locator("article").filter({ hasText: email });
  await memberRow.hover();
  await memberRow
    .getByRole("button", { name: "Remove member from project" })
    .click();
  const removeDialog = ownerPage.getByRole("dialog", {
    name: "Remove member from project",
  });
  await removeDialog.getByRole("button", { name: "Remove" }).click();
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
    memberPage.getByRole("heading", { name: projectName, exact: true }),
  ).toBeVisible();
  await expect(
    ownerPage.locator("div.border-green-500").filter({ hasText: "R" }),
  ).toBeVisible({ timeout: 15_000 });

  await ownerPage.getByRole("button", { name: "Settings" }).click();
  const settings = ownerPage.getByRole("dialog", { name: "Project settings" });
  await settings.locator("#name").fill(renamedProject);
  await settings.locator("#column-0").fill(renamedColumn);
  await settings
    .locator("#column-description-0")
    .fill("Updated while another member watches the board.");
  await settings.getByRole("button", { name: "Move Done up" }).click();
  await expect(columnEditor(settings, 1).locator("#column-1")).toHaveValue("Done");
  await settings.getByRole("button", { name: "Save changes" }).click();
  await expectToast(ownerPage, "Project saved successfully");

  await expect(
    memberPage.getByRole("heading", { name: renamedProject, exact: true }),
  ).toBeVisible({ timeout: 15_000 });
  await expect(
    memberPage.getByRole("heading", { level: 3 }),
  ).toHaveText([renamedColumn, "Done", "Doing"]);
  await expect(memberPage.getByRole("heading", { name: projectName })).toHaveCount(0);

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

  await expect(projectListLink(memberPage, projectName)).toBeVisible({ timeout: 15_000 });
  await memberPage.getByRole("button", { name: "Notifications" }).click();
  await expect(
    memberPage
      .getByRole("button")
      .filter({ hasText: `${testUser.name} added you to ${projectName}.` }),
  ).toBeVisible({ timeout: 15_000 });
  await memberPage.keyboard.press("Escape");

  await projectListLink(memberPage, projectName).click();
  await memberPage.getByTitle("View project members").click();
  const membersDialog = memberPage.getByRole("dialog", { name: "Project members" });
  await expect(membersDialog.getByText("2 members")).toBeVisible();
  await expect(membersDialog.getByText(member.email)).toBeVisible();
  await memberPage.keyboard.press("Escape");

  await memberPage.getByRole("link", { name: "Chat" }).click();
  await expect(memberPage.getByPlaceholder("Message...")).toBeEnabled({ timeout: 15_000 });
  await memberPage.getByRole("link", { name: "Go back" }).click();
  await expect(
    ownerPage.locator("div.border-green-500").filter({ hasText: "R" }),
  ).toBeVisible({ timeout: 15_000 });

  await createTask(ownerPage, taskTitle);
  await expect(memberPage.getByText(taskTitle, { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await expect(
    ownerPage.locator("div.border-green-500").filter({ hasText: "R" }),
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
  await expect(projectListLink(memberPage, projectName)).toBeVisible({ timeout: 15_000 });
  await projectListLink(memberPage, projectName).click();

  const task = await createTask(ownerPage, taskTitle);
  await expect(memberPage.getByText(taskTitle, { exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await memberPage.getByRole("link", { name: "Chat" }).click();
  await expect(memberPage.getByPlaceholder("Message...")).toBeEnabled({ timeout: 15_000 });
  await memberPage.getByRole("link", { name: "Go back" }).click();
  await expect(
    ownerPage.locator("div.border-green-500").filter({ hasText: "R" }),
  ).toBeVisible({ timeout: 15_000 });

  await removeMember(ownerPage, member.email);

  await expect(memberPage).toHaveURL(/\/projects\/?$/, { timeout: 15_000 });
  await expect(projectListLink(memberPage, projectName)).toHaveCount(0);
  await expect(
    memberPage.getByRole("heading", { name: projectName, exact: true }),
  ).toHaveCount(0);

  const chatResponsePromise = memberPage.waitForResponse(
    (response) =>
      response.url().startsWith(backendURL) &&
      new URL(response.url()).pathname === `/projects/${projectId}/chat`,
  );
  await memberPage.goto(`/projects/${projectId}/chat`);
  expect((await chatResponsePromise).status()).toBe(403);

  const taskResponsePromise = memberPage.waitForResponse(
    (response) =>
      response.url().startsWith(backendURL) &&
      new URL(response.url()).pathname === `/tasks/${task.id}`,
  );
  const projectResponsePromise = memberPage.waitForResponse(
    (response) =>
      response.url().startsWith(backendURL) &&
      new URL(response.url()).pathname === `/projects/${projectId}`,
  );
  await memberPage.goto(`/projects/${projectId}?taskId=${task.id}`);
  expect((await taskResponsePromise).status()).toBe(403);
  expect((await projectResponsePromise).status()).toBe(403);

  await memberPage.context().close();
});

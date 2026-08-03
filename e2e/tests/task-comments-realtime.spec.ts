import crypto from "node:crypto";
import {
  addProjectMember,
  createProjectThroughUI,
  expect,
  expectToast,
  loginAsUser,
  test,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";
import type { Locator, Page } from "@playwright/test";

async function fillRichText(editor: Locator, page: Page, value: string) {
  await editor.click();
  await page.keyboard.press(
    process.platform === "darwin" ? "Meta+A" : "Control+A"
  );
  await editor.pressSequentially(value);
}

test("project collaborators receive new task comments in an open task in real time", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "Realtime Commenter",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Realtime Comments Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Realtime Comments Task ${crypto.randomUUID()}`;
  const commentText = `E2E live comment ${crypto.randomUUID()}`;

  await createProjectThroughUI(ownerPage, projectName, {
    description: "Created by the real-time task comments e2e test.",
  });

  await ownerPage.getByRole("link", { name: projectName }).click();
  const projectId = ownerPage.url().split("/projects/")[1];
  await addProjectMember(ownerPage, member.email);

  await ownerPage.reload();
  await ownerPage
    .getByRole("button", { name: "Open actions for Doing" })
    .click();
  await ownerPage.getByRole("menuitem", { name: "Add task" }).click();
  const createTaskDialog = ownerPage.getByRole("dialog", {
    name: "Create task",
  });
  await createTaskDialog.locator("#title").fill(taskTitle);
  await fillRichText(
    createTaskDialog.locator("#description"),
    ownerPage,
    "A task with a live collaborator comment feed."
  );
  await createTaskDialog.locator("#priority").click();
  await ownerPage.getByRole("option", { name: "Medium", exact: true }).click();
  await createTaskDialog.getByRole("button", { name: "Create task" }).click();
  await expectToast(ownerPage, "Task created successfully");

  await memberPage.goto(`/projects/${projectId}`);
  await expect(memberPage.getByText(taskTitle, { exact: true })).toBeVisible();
  await expect(
    ownerPage.locator("div.border-success").filter({ hasText: "R" })
  ).toBeVisible({ timeout: 15_000 });

  await ownerPage.getByText(taskTitle, { exact: true }).click();
  const ownerTaskDetails = ownerPage.getByRole("dialog", { name: taskTitle });
  await expect(ownerTaskDetails).toBeVisible();

  await memberPage.getByText(taskTitle, { exact: true }).click();
  const memberTaskDetails = memberPage.getByRole("dialog", { name: taskTitle });
  await expect(memberTaskDetails).toBeVisible();

  const commentEditor = ownerTaskDetails.locator(
    '[aria-placeholder="Share context, decisions, or blockers..."]'
  );
  await fillRichText(commentEditor, ownerPage, commentText);
  await ownerTaskDetails.getByRole("button", { name: "Comment" }).click();
  await expect(
    ownerTaskDetails.getByText(commentText, { exact: true })
  ).toBeVisible();

  await expect(
    memberTaskDetails.getByText(commentText, { exact: true })
  ).toBeVisible({ timeout: 15_000 });

  await memberPage.context().close();
});

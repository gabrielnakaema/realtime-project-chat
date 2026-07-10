import crypto from "node:crypto";
import {
  expect,
  expectToast,
  loginAsUser,
  test,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";

test("assigned project collaborator receives a notification and can open the task", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "E2E Notification Member",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Notification Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Assigned Task ${crypto.randomUUID()}`;

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
    .pressSequentially("Created by the notification e2e test");
  await createProjectDialog
    .getByRole("button", { name: "Create project" })
    .click();
  await expectToast(ownerPage, "Project created successfully");

  await ownerPage.getByRole("link", { name: projectName }).click();
  await ownerPage.getByTitle("Add project member").click();
  const addMemberDialog = ownerPage.getByRole("dialog", {
    name: "Add project member",
  });
  await addMemberDialog.getByLabel("Email").fill(member.email);
  await addMemberDialog.getByRole("button", { name: "Add member" }).click();
  await expectToast(ownerPage, "Member added successfully");

  await ownerPage.reload();
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
    .pressSequentially("A task that notifies its assigned collaborator.");
  await createTaskDialog.locator("#responsible_id").click();
  await ownerPage
    .getByRole("option", { name: member.name, exact: true })
    .click();
  await createTaskDialog.locator("#priority").click();
  await ownerPage.getByRole("option", { name: "Medium", exact: true }).click();
  await createTaskDialog.getByRole("button", { name: "Create task" }).click();
  await expectToast(ownerPage, "Task created successfully");

  await memberPage.getByRole("button", { name: "Notifications" }).click();
  const assignmentNotification = memberPage
    .getByRole("button")
    .filter({ hasText: `assigned you to ${taskTitle}` });
  await expect(assignmentNotification).toBeVisible({ timeout: 15_000 });

  await assignmentNotification.click();
  await expect(
    memberPage.getByRole("dialog", { name: taskTitle })
  ).toBeVisible();
  await expect(memberPage).toHaveURL(/\/projects\/[^/]+\?taskId=/);

  await memberPage.context().close();
});

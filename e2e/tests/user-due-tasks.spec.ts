import crypto from "node:crypto";
import {
  expect,
  expectToast,
  loginAsUser,
  test,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";

test("assigned user can open an upcoming task from the dashboard", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "E2E Due Task Member",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Due Task Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Due Task ${crypto.randomUUID()}`;
  const dueDate = "2026-08-20";

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
    .pressSequentially("Created by the user due-tasks e2e test");
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
    .pressSequentially("A due task that appears on the assignee dashboard.");
  await createTaskDialog.locator("#responsible_id").click();
  await ownerPage
    .getByRole("option", { name: member.name, exact: true })
    .click();
  await createTaskDialog.locator("#priority").click();
  await ownerPage.getByRole("option", { name: "High", exact: true }).click();
  await createTaskDialog.locator("#due_date").fill(dueDate);
  await createTaskDialog.getByRole("button", { name: "Create task" }).click();
  await expectToast(ownerPage, "Task created successfully");

  await memberPage.reload();
  const dueTasks = memberPage
    .getByRole("heading", {
      name: "Upcoming deadlines",
      exact: true,
    })
    .locator("xpath=ancestor::section[1]");
  await expect(dueTasks.getByText(taskTitle, { exact: true })).toBeVisible();
  await expect(dueTasks.getByText("Aug 20", { exact: true })).toBeVisible();
  await dueTasks
    .getByRole("button", { name: `Open task ${taskTitle}`, exact: true })
    .click();

  const taskDetails = memberPage.getByRole("dialog", { name: taskTitle });
  await expect(taskDetails).toBeVisible();

  await memberPage.context().close();
});

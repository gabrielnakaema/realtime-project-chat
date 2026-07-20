import crypto from "node:crypto";
import {
  addProjectMember,
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

  await ownerPage.getByRole("button", { name: "New project" }).first().click();

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
  const dueTasks = memberPage.getByRole("region", {
    name: "Upcoming Deadlines",
  });
  const dueTaskLink = dueTasks.getByRole("link", {
    name: new RegExp(taskTitle),
  });
  await expect(dueTaskLink).toBeVisible();
  await expect(dueTaskLink.getByText(/Due (in|today|tomorrow)/)).toBeVisible();
  await dueTaskLink.click();

  const taskDetails = memberPage.getByRole("dialog", { name: taskTitle });
  await expect(taskDetails).toBeVisible();

  await memberPage.context().close();
});

test("assigned user can complete an upcoming task and remove it from the dashboard", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "E2E Completing Due Task Member",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Complete Due Task Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Complete Due Task ${crypto.randomUUID()}`;
  const dueDate = "2026-08-20";

  await ownerPage.getByRole("button", { name: "New project" }).first().click();
  const createProjectDialog = ownerPage.getByRole("dialog", {
    name: "Create project",
  });
  await createProjectDialog.locator("#name").fill(projectName);
  await createProjectDialog
    .locator("#description")
    .pressSequentially("Created by the due-task completion e2e test.");
  await createProjectDialog
    .getByRole("button", { name: "Create project" })
    .click();
  await expectToast(ownerPage, "Project created successfully");

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
  await createTaskDialog
    .locator("#description")
    .pressSequentially("An assignee completes this upcoming task.");
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
  const dueTasks = memberPage.getByRole("region", {
    name: "Upcoming Deadlines",
  });
  await expect(
    dueTasks.getByRole("link", { name: new RegExp(taskTitle) })
  ).toBeVisible();

  await memberPage.goto(`/projects/${projectId}`);
  const draggableTask = memberPage
    .getByText(taskTitle, { exact: true })
    .locator("xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]");
  const doneColumn = memberPage
    .getByRole("button", { name: "Open actions for Done" })
    .locator(
      "xpath=ancestor::div[.//div[contains(@class, 'overflow-y-auto')]][1]"
    );
  const moveResponsePromise = memberPage.waitForResponse((response) => {
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
  await expect(doneColumn.getByText(taskTitle, { exact: true })).toBeVisible();

  await memberPage.goto("/projects");
  const updatedDueTasks = memberPage.getByRole("region", {
    name: "Upcoming Deadlines",
  });
  await expect(
    updatedDueTasks.getByRole("link", { name: new RegExp(taskTitle) })
  ).toHaveCount(0);
  await expect(
    updatedDueTasks.getByText("Nothing due soon", { exact: true })
  ).toBeVisible();

  await memberPage.context().close();
});

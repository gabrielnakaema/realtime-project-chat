import crypto from "node:crypto";
import {
  addProjectMember,
  expect,
  expectToast,
  loginAsUser,
  test,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";

test("new project member receives a notification and can open the project", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
  testUser,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "E2E Member Notification Recipient",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Member Notification Project ${crypto.randomUUID()}`;

  await ownerPage.getByRole("button", { name: "New project" }).first().click();
  const createProjectDialog = ownerPage.getByRole("dialog", {
    name: "Create project",
  });
  await createProjectDialog.locator("#name").fill(projectName);
  await createProjectDialog
    .locator("#description")
    .pressSequentially("Created by the project-member notification e2e test.");
  await createProjectDialog
    .getByRole("button", { name: "Create project" })
    .click();
  await expectToast(ownerPage, "Project created successfully");

  await ownerPage.getByRole("link", { name: projectName }).click();
  await addProjectMember(ownerPage, member.email);

  await memberPage.getByRole("button", { name: "Notifications" }).click();
  const projectMemberNotification = memberPage
    .getByRole("button")
    .filter({ hasText: `${testUser.name} added you to ${projectName}.` });
  await expect(projectMemberNotification).toBeVisible({ timeout: 15_000 });

  await projectMemberNotification.click();
  await expect(memberPage).toHaveURL(/\/projects\/[^/?]+$/);
  await expect(
    memberPage.getByRole("button", { name: projectName, exact: true })
  ).toBeVisible();

  await memberPage.context().close();
});

test("user can mark all project notifications as read and the state persists", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
  testUser,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "E2E Notification Inbox Member",
  });
  const memberPage = await loginAsUser(browser, member);
  const firstProjectName = `E2E Notification Inbox One ${crypto.randomUUID()}`;
  const secondProjectName = `E2E Notification Inbox Two ${crypto.randomUUID()}`;

  const addMemberToNewProject = async (projectName: string) => {
    await ownerPage
      .getByRole("button", { name: "New project" })
      .first()
      .click();
    const createProjectDialog = ownerPage.getByRole("dialog", {
      name: "Create project",
    });
    await createProjectDialog.locator("#name").fill(projectName);
    await createProjectDialog
      .locator("#description")
      .pressSequentially("Created by the notification inbox e2e test.");
    await createProjectDialog
      .getByRole("button", { name: "Create project" })
      .click();
    await expectToast(ownerPage, "Project created successfully");

    await ownerPage.getByRole("link", { name: projectName }).click();
    await addProjectMember(ownerPage, member.email);
    await ownerPage.goto("/projects");
  };

  await addMemberToNewProject(firstProjectName);
  await addMemberToNewProject(secondProjectName);

  await memberPage.getByRole("button", { name: "Notifications" }).click();
  const firstNotification = memberPage
    .getByRole("button")
    .filter({ hasText: `${testUser.name} added you to ${firstProjectName}.` });
  const secondNotification = memberPage
    .getByRole("button")
    .filter({ hasText: `${testUser.name} added you to ${secondProjectName}.` });
  await expect(firstNotification).toBeVisible({ timeout: 15_000 });
  await expect(secondNotification).toBeVisible({ timeout: 15_000 });
  await expect(firstNotification.locator("span.bg-primary")).toHaveCount(1);
  await expect(secondNotification.locator("span.bg-primary")).toHaveCount(1);

  const markAllReadButton = memberPage.getByRole("button", {
    name: "Mark all read",
  });
  const markAllReadResponsePromise = memberPage.waitForResponse((response) => {
    const url = new URL(response.url());

    return (
      response.request().method() === "POST" &&
      url.pathname.endsWith("/notifications/read-all")
    );
  });
  await markAllReadButton.click();
  expect((await markAllReadResponsePromise).ok()).toBe(true);
  await expect(markAllReadButton).toBeDisabled();
  await expect(firstNotification.locator("span.bg-primary")).toHaveCount(0);
  await expect(secondNotification.locator("span.bg-primary")).toHaveCount(0);

  await memberPage.reload();
  await memberPage.getByRole("button", { name: "Notifications" }).click();
  await expect(
    memberPage.getByRole("button", { name: "Mark all read" })
  ).toBeDisabled();
  await expect(
    memberPage
      .getByRole("button")
      .filter({ hasText: `${testUser.name} added you to ${firstProjectName}.` })
      .locator("span.bg-primary")
  ).toHaveCount(0);
  await expect(
    memberPage
      .getByRole("button")
      .filter({
        hasText: `${testUser.name} added you to ${secondProjectName}.`,
      })
      .locator("span.bg-primary")
  ).toHaveCount(0);

  await memberPage.context().close();
});

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

  await ownerPage.getByRole("button", { name: "New project" }).first().click();

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

test("task author receives a comment notification and opens the commented task", async ({
  authenticatedPage: ownerPage,
  backendURL,
  browser,
  request,
}) => {
  const member = await registerUser(request, backendURL, {
    name: "Realtime Commenter",
  });
  const memberPage = await loginAsUser(browser, member);
  const projectName = `E2E Comment Notification Project ${crypto.randomUUID()}`;
  const taskTitle = `E2E Commented Task ${crypto.randomUUID()}`;
  const commentText = `E2E notification comment ${crypto.randomUUID()}`;

  await ownerPage.getByRole("button", { name: "New project" }).first().click();
  const createProjectDialog = ownerPage.getByRole("dialog", {
    name: "Create project",
  });
  await createProjectDialog.locator("#name").fill(projectName);
  await createProjectDialog
    .locator("#description")
    .pressSequentially("Created by the comment notification e2e test.");
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
    .pressSequentially(
      "A task whose author should be notified about comments."
    );
  await createTaskDialog.locator("#priority").click();
  await ownerPage.getByRole("option", { name: "High", exact: true }).click();
  await createTaskDialog.getByRole("button", { name: "Create task" }).click();
  await expectToast(ownerPage, "Task created successfully");

  await ownerPage.goto("/projects");
  await memberPage.goto(`/projects/${projectId}`);
  await memberPage.getByText(taskTitle, { exact: true }).click();
  const memberTaskDetails = memberPage.getByRole("dialog", { name: taskTitle });
  const commentEditor = memberTaskDetails.locator(
    '[aria-placeholder="Share context, decisions, or blockers..."]'
  );
  await commentEditor.click();
  await commentEditor.pressSequentially(commentText);
  await memberTaskDetails.getByRole("button", { name: "Comment" }).click();
  await expect(
    memberTaskDetails.getByText(commentText, { exact: true })
  ).toBeVisible();

  await ownerPage.getByRole("button", { name: "Notifications" }).click();
  const commentNotification = ownerPage
    .getByRole("button")
    .filter({ hasText: `${member.name} commented on ${taskTitle}` });
  await expect(commentNotification).toBeVisible({ timeout: 15_000 });

  await commentNotification.click();
  const ownerTaskDetails = ownerPage.getByRole("dialog", { name: taskTitle });
  await expect(ownerTaskDetails).toBeVisible();
  await expect(
    ownerTaskDetails.getByText(commentText, { exact: true })
  ).toBeVisible();
  await expect(ownerPage).toHaveURL(/\/projects\/[^/]+\?taskId=.*commentId=/);

  await memberPage.context().close();
});

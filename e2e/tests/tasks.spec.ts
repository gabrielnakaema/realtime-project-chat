import crypto from "node:crypto";
import {
  createProjectThroughUI,
  expect,
  expectToast,
  test,
} from "../src/fixtures/authenticated-page.js";
import { fillRichTextList } from "../src/fixtures/rich-text-editor.js";
import type { Locator, Page } from "@playwright/test";

async function fillRichText(editor: Locator, page: Page, value: string) {
  await editor.click();
  await page.keyboard.press(
    process.platform === "darwin" ? "Meta+A" : "Control+A"
  );
  await editor.pressSequentially(value);
}

async function createProject(page: Page, name: string) {
  await createProjectThroughUI(page, name, {
    description: "Project created by the task creation e2e test",
  });
  await page.getByRole("link", { name: new RegExp(name) }).click();
  await expect(page.getByRole("button", { name, exact: true })).toBeVisible();
}

async function selectOption(
  container: Locator,
  page: Page,
  fieldId: string,
  optionName: string
) {
  await container.locator(`#${fieldId}`).click();
  await page.getByRole("option", { name: optionName, exact: true }).click();
}

function taskCard(page: Page, title: string) {
  return page.getByText(title, { exact: true });
}

function headerCreateTaskButton(page: Page) {
  return page.locator("header").getByRole("button", { name: "Create task" });
}

function boardColumn(page: Page, columnName: string) {
  return page
    .getByRole("button", { name: `Open actions for ${columnName}` })
    .locator(
      "xpath=ancestor::div[.//div[contains(@class, 'overflow-y-auto')]][1]"
    );
}

function taskCardInColumn(page: Page, columnName: string, title: string) {
  return boardColumn(page, columnName).getByText(title, { exact: true });
}

async function openTaskDetails(page: Page, title: string) {
  await taskCard(page, title).click();
  return page.getByRole("dialog", { name: title });
}

async function dragTaskToColumn(
  page: Page,
  taskTitle: string,
  columnName: string
) {
  const draggableTask = taskCard(page, taskTitle).locator(
    "xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]"
  );
  const targetColumn = boardColumn(page, columnName);

  const moveResponsePromise = page.waitForResponse((response) => {
    const url = new URL(response.url());

    return (
      response.request().method() === "PATCH" &&
      /\/tasks\/[^/]+\/move$/.test(url.pathname)
    );
  });

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
  const draggableTask = taskCard(page, taskTitle).locator(
    "xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]"
  );
  const targetTask = taskCard(page, targetTaskTitle).locator(
    "xpath=ancestor::div[contains(@class, 'cursor-pointer')][1]"
  );

  const moveResponsePromise = page.waitForResponse((response) => {
    const url = new URL(response.url());

    return (
      response.request().method() === "PATCH" &&
      /\/tasks\/[^/]+\/move$/.test(url.pathname)
    );
  });

  await draggableTask.dragTo(targetTask, {
    targetPosition: { x: 40, y: 1 },
  });

  const moveResponse = await moveResponsePromise;
  expect(moveResponse.ok()).toBe(true);

  return moveResponse;
}

test.describe("tasks", () => {
  test("project owner can create a task in a board column and view its details", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Project ${crypto.randomUUID()}`;
    const taskTitle = `E2E Task ${crypto.randomUUID()}`;
    const taskDescription = "A task created from the Doing column.";

    await createProject(page, projectName);

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      taskDescription
    );
    await selectOption(createDialog, page, "priority", "High");
    await createDialog.getByRole("button", { name: "Create task" }).click();

    await expectToast(page, "Task created successfully");

    const taskCard = page.getByText(taskTitle, { exact: true });
    await expect(taskCard).toBeVisible();
    await taskCard.click();

    const taskDetails = page.getByRole("dialog", { name: taskTitle });
    await expect(
      taskDetails.getByText(taskDescription, { exact: true })
    ).toBeVisible();
    await expect(
      taskDetails
        .getByText("Column", { exact: true })
        .locator("..")
        .getByText("Doing", { exact: true })
    ).toBeVisible();
    await expect(
      taskDetails
        .getByText("Priority", { exact: true })
        .locator("..")
        .getByText("High", { exact: true })
    ).toBeVisible();
  });

  test("project owner can select a suggested task code when creating a task", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Code Project ${crypto.randomUUID()}`;
    const codePrefix = `E2E-${crypto.randomUUID().slice(0, 8).toUpperCase()}-`;
    const existingTaskTitle = `E2E Existing Code Task ${crypto.randomUUID()}`;
    const taskTitle = `E2E Suggested Code Task ${crypto.randomUUID()}`;
    const existingCode = `${codePrefix}001`;
    const suggestedCode = `${codePrefix}002`;

    await createProject(page, projectName);

    await headerCreateTaskButton(page).click();
    let createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(existingTaskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "Creates the first task code in the sequence."
    );
    await createDialog.getByPlaceholder("TASK-101").fill(existingCode);
    await selectOption(createDialog, page, "priority", "Medium");
    await createDialog.getByRole("button", { name: "Create task" }).click();
    await expectToast(page, "Task created successfully");

    await headerCreateTaskButton(page).click();
    createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "Uses the next task code suggested by the project."
    );

    const codeInput = createDialog.getByPlaceholder("TASK-101");
    await codeInput.fill(codePrefix);
    await page
      .getByRole("option", { name: new RegExp(`^${suggestedCode}`) })
      .click();
    await expect(codeInput).toHaveValue(suggestedCode);

    await selectOption(createDialog, page, "priority", "High");
    await createDialog.getByRole("button", { name: "Create task" }).click();
    await expectToast(page, "Task created successfully");

    let taskDetails = await openTaskDetails(page, taskTitle);
    await expect(taskDetails.getByTitle(suggestedCode).first()).toBeVisible();

    await page.keyboard.press("Escape");
    await page.reload();
    taskDetails = await openTaskDetails(page, taskTitle);
    await expect(taskDetails.getByTitle(suggestedCode).first()).toBeVisible();
  });

  test("project owner can move a task across board columns and the move persists", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Move Project ${crypto.randomUUID()}`;
    const taskTitle = `E2E Moving Task ${crypto.randomUUID()}`;

    await createProject(page, projectName);

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "A task moved through the board workflow."
    );
    await selectOption(createDialog, page, "priority", "Medium");
    await createDialog.getByRole("button", { name: "Create task" }).click();

    await expectToast(page, "Task created successfully");
    await expect(taskCard(page, taskTitle)).toBeVisible();

    await dragTaskToColumn(page, taskTitle, "Done");

    await expect(taskCardInColumn(page, "Doing", taskTitle)).toHaveCount(0);
    await expect(taskCardInColumn(page, "Done", taskTitle)).toBeVisible();
    await taskCardInColumn(page, "Done", taskTitle).click();

    let taskDetails = page.getByRole("dialog", { name: taskTitle });
    await expect(
      taskDetails
        .getByText("Column", { exact: true })
        .locator("..")
        .getByText("Done", { exact: true })
    ).toBeVisible();

    await page.keyboard.press("Escape");
    await expect(taskDetails).toHaveCount(0);
    await page.reload();

    await expect(taskCardInColumn(page, "Done", taskTitle)).toBeVisible();
    await taskCardInColumn(page, "Done", taskTitle).click();
    taskDetails = page.getByRole("dialog", { name: taskTitle });
    await expect(
      taskDetails
        .getByText("Column", { exact: true })
        .locator("..")
        .getByText("Done", { exact: true })
    ).toBeVisible();
  });

  test("project owner can reorder tasks within a board column and the order persists", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Order Project ${crypto.randomUUID()}`;
    const firstTaskTitle = `E2E First Task ${crypto.randomUUID()}`;
    const secondTaskTitle = `E2E Second Task ${crypto.randomUUID()}`;
    const thirdTaskTitle = `E2E Third Task ${crypto.randomUUID()}`;

    await createProject(page, projectName);

    const createTask = async (title: string) => {
      await headerCreateTaskButton(page).click();
      const createDialog = page.getByRole("dialog", { name: "Create task" });
      await createDialog.locator("#title").fill(title);
      await fillRichText(
        createDialog.locator("#description"),
        page,
        "A task used to verify the board order."
      );
      await selectOption(createDialog, page, "priority", "Medium");
      await createDialog.getByRole("button", { name: "Create task" }).click();
      await expect(createDialog).toHaveCount(0);
      await expect(taskCardInColumn(page, "Pending", title)).toBeVisible();
    };

    await createTask(firstTaskTitle);
    await createTask(secondTaskTitle);
    await createTask(thirdTaskTitle);

    const pendingTaskHeadings = boardColumn(page, "Pending").getByRole(
      "heading",
      { level: 4 }
    );
    await expect(pendingTaskHeadings).toHaveText([
      thirdTaskTitle,
      secondTaskTitle,
      firstTaskTitle,
    ]);

    await dragTaskBeforeTask(page, firstTaskTitle, thirdTaskTitle);
    await expect(pendingTaskHeadings).toHaveText([
      firstTaskTitle,
      thirdTaskTitle,
      secondTaskTitle,
    ]);

    await page.reload();
    await expect(
      boardColumn(page, "Pending").getByRole("heading", { level: 4 })
    ).toHaveText([firstTaskTitle, thirdTaskTitle, secondTaskTitle]);
  });

  test("task creation form shows validation errors for missing required fields", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Validation Project ${crypto.randomUUID()}`;

    await createProject(page, projectName);

    await headerCreateTaskButton(page).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.getByRole("button", { name: "Create task" }).click();

    await expect(createDialog.getByText("Title is required")).toBeVisible();
    await expect(createDialog.getByText("Priority is required")).toBeVisible();
    await expect(createDialog).toBeVisible();
  });

  test("task descriptions preserve bulleted lists through create, edit, and reload", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task List Project ${crypto.randomUUID()}`;
    const taskTitle = `E2E Task List ${crypto.randomUUID()}`;
    const listItems = ["Plan the change", "Verify the result"];
    const addedItem = "Document the behavior";

    await createProject(page, projectName);
    await headerCreateTaskButton(page).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichTextList(createDialog, "bulleted", listItems);
    await selectOption(createDialog, page, "priority", "Medium");
    await createDialog.getByRole("button", { name: "Create task" }).click();

    await expectToast(page, "Task created successfully");

    let taskDetails = await openTaskDetails(page, taskTitle);
    let renderedList = taskDetails.locator("ul").filter({
      hasText: listItems[0],
    });
    await expect(renderedList.locator("li")).toHaveText(listItems);
    await expect(renderedList).toHaveCSS("list-style-type", "disc");

    await taskDetails.getByRole("button", { name: "Edit task" }).click();
    const editDialog = page.getByRole("dialog", { name: "Edit task" });
    const editorList = editDialog.locator("#description ul");
    await expect(editorList.locator("li")).toHaveText(listItems);

    await editorList.locator("li").last().click();
    await page.keyboard.press("End");
    await page.keyboard.press("Enter");
    await page.keyboard.type(addedItem);
    await editDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Task updated successfully");
    taskDetails = page.getByRole("dialog", { name: taskTitle });
    renderedList = taskDetails.locator("ul").filter({
      hasText: listItems[0],
    });
    await expect(renderedList.locator("li")).toHaveText([
      ...listItems,
      addedItem,
    ]);

    await page.keyboard.press("Escape");
    await page.reload();
    taskDetails = await openTaskDetails(page, taskTitle);
    renderedList = taskDetails.locator("ul").filter({
      hasText: listItems[0],
    });
    await expect(renderedList.locator("li")).toHaveText([
      ...listItems,
      addedItem,
    ]);
    await expect(renderedList).toHaveCSS("list-style-type", "disc");
  });

  test("project owner can create a task with all optional fields, edit it, and see the activity timeline update", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Full Fields Project ${crypto.randomUUID()}`;
    const taskTitle = `E2E Full Task ${crypto.randomUUID()}`;
    const taskCode = `TSK-${crypto.randomUUID().slice(0, 8)}`;
    const dueDate = "2026-08-20";

    await createProject(page, projectName);

    await headerCreateTaskButton(page).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "A task created with every optional field filled in."
    );
    await createDialog.getByPlaceholder("TASK-101").fill(taskCode);
    await selectOption(createDialog, page, "responsible_id", "E2E User");
    await selectOption(createDialog, page, "priority", "Medium");
    await createDialog.locator("#due_date").fill(dueDate);
    await createDialog.locator("#tags").fill("backend, urgent");
    await createDialog.getByRole("button", { name: "Create task" }).click();

    await expectToast(page, "Task created successfully");

    let taskDetails = await openTaskDetails(page, taskTitle);
    await expect(taskDetails.getByTitle(taskCode).first()).toBeVisible();
    await expect(
      taskDetails
        .getByText("Responsible", { exact: true })
        .locator("..")
        .getByText("E2E User", { exact: true })
    ).toBeVisible();
    await expect(
      taskDetails.getByText("backend", { exact: true })
    ).toBeVisible();
    await expect(
      taskDetails.getByText("urgent", { exact: true })
    ).toBeVisible();

    await taskDetails.getByRole("button", { name: "Edit task" }).click();

    const editDialog = page.getByRole("dialog", { name: "Edit task" });
    await expect(editDialog.locator("#title")).toHaveValue(taskTitle);
    await expect(editDialog.getByPlaceholder("TASK-101")).toHaveValue(taskCode);
    await expect(editDialog.locator("#due_date")).toHaveValue(dueDate);
    await expect(editDialog.locator("#tags")).toHaveValue("backend,urgent");

    await selectOption(editDialog, page, "priority", "Low");
    await editDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Task updated successfully");

    taskDetails = page.getByRole("dialog", { name: taskTitle });
    await expect(
      taskDetails
        .getByText("Priority", { exact: true })
        .locator("..")
        .getByText("Low", { exact: true })
    ).toBeVisible();

    await expect(async () => {
      await page.reload();
      taskDetails = page.getByRole("dialog", { name: taskTitle });
      await expect(taskDetails).toBeVisible();
      await taskDetails
        .getByRole("button", { name: "Activity timeline" })
        .click();
      await expect(taskDetails.getByText("set priority to")).toBeVisible({
        timeout: 1_000,
      });
    }).toPass({ timeout: 15_000 });

    await expect(taskDetails.getByText("created the task")).toBeVisible();
  });

  test("project owner can cancel and then confirm archiving a task, and restore it from the archived tasks list", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Archive Project ${crypto.randomUUID()}`;
    const taskTitle = `E2E Archive Task ${crypto.randomUUID()}`;

    await createProject(page, projectName);

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "A task created to test the archive and restore flow."
    );
    await selectOption(createDialog, page, "priority", "Low");
    await createDialog.getByRole("button", { name: "Create task" }).click();
    await expectToast(page, "Task created successfully");

    let taskDetails = await openTaskDetails(page, taskTitle);

    await taskDetails.getByRole("button", { name: "Archive task" }).click();
    await expect(taskDetails.getByText("Archive task?")).toBeVisible();
    await taskDetails.getByRole("button", { name: "Cancel" }).click();
    await expect(taskDetails.getByText("Archive task?")).toHaveCount(0);
    await expect(
      taskDetails.getByRole("button", { name: "Archive task" })
    ).toBeVisible();

    await taskDetails.getByRole("button", { name: "Archive task" }).click();
    await taskDetails.getByRole("button", { name: "Confirm" }).click();

    await expect(taskDetails).toHaveCount(0);
    await expect(taskCard(page, taskTitle)).toHaveCount(0);

    await page.getByRole("button", { name: "Archived" }).click();
    const archivedDialog = page.getByRole("dialog", {
      name: "Archived tasks",
    });
    await expect(
      archivedDialog.getByText(taskTitle, { exact: true })
    ).toBeVisible();

    await archivedDialog.getByTitle("Restore task").click();
    await archivedDialog.getByRole("button", { name: "Doing" }).click();

    await expect(
      archivedDialog.getByText(taskTitle, { exact: true })
    ).toHaveCount(0);
    await page.keyboard.press("Escape");
    await expect(archivedDialog).toHaveCount(0);

    await expect(taskCard(page, taskTitle)).toBeVisible();
    taskDetails = await openTaskDetails(page, taskTitle);
    await expect(
      taskDetails
        .getByText("Column", { exact: true })
        .locator("..")
        .getByText("Doing", { exact: true })
    ).toBeVisible();
  });

  test("users can post a comment and reply to it on a task", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Comments Project ${crypto.randomUUID()}`;
    const taskTitle = `E2E Comment Task ${crypto.randomUUID()}`;
    const commentText = `Initial comment ${crypto.randomUUID()}`;
    const replyText = `Reply to that comment ${crypto.randomUUID()}`;

    await createProject(page, projectName);

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "A task created to test comments and replies."
    );
    await selectOption(createDialog, page, "priority", "Low");
    await createDialog.getByRole("button", { name: "Create task" }).click();
    await expectToast(page, "Task created successfully");

    const taskDetails = await openTaskDetails(page, taskTitle);

    const commentEditor = taskDetails.locator(
      '[aria-placeholder="Share context, decisions, or blockers..."]'
    );
    await fillRichText(commentEditor, page, commentText);
    await taskDetails.getByRole("button", { name: "Comment" }).click();

    const commentArticle = taskDetails
      .locator("article")
      .filter({ hasText: commentText });
    await expect(commentArticle).toBeVisible();
    await expect(commentArticle.getByText("You")).toBeVisible();

    await commentArticle.getByRole("button", { name: "Reply" }).click();

    const replyEditor = taskDetails.locator(
      '[aria-placeholder="Reply to E2E User..."]'
    );
    await fillRichText(replyEditor, page, replyText);
    await commentArticle.getByRole("button", { name: "Reply" }).last().click();

    await expect(
      taskDetails.locator("article").filter({ hasText: replyText })
    ).toBeVisible();
  });

  test("project owner can link a task dependency and navigate to it from task details", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Dependency Project ${crypto.randomUUID()}`;
    const upstreamTaskTitle = `E2E Upstream Task ${crypto.randomUUID()}`;
    const downstreamTaskTitle = `E2E Downstream Task ${crypto.randomUUID()}`;

    await createProject(page, projectName);

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    let createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(upstreamTaskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "The task that the downstream task depends on."
    );
    await selectOption(createDialog, page, "priority", "Low");
    await createDialog.getByRole("button", { name: "Create task" }).click();
    await expectToast(page, "Task created successfully");

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(downstreamTaskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "The task that depends on the upstream task."
    );
    await selectOption(createDialog, page, "priority", "Low");

    await createDialog
      .getByRole("button", { name: "Search tasks to depend on" })
      .click();
    await page
      .getByPlaceholder("Search by code, title, id, or description...")
      .fill(upstreamTaskTitle);
    await page
      .getByRole("option", { name: new RegExp(upstreamTaskTitle) })
      .click();

    await page.keyboard.press("Escape");
    await expect(
      createDialog.getByRole("button", {
        name: `Remove dependency ${upstreamTaskTitle}`,
      })
    ).toBeVisible();

    await createDialog.getByRole("button", { name: "Create task" }).click();
    await expectToast(page, "Task created successfully");

    const downstreamDetails = await openTaskDetails(page, downstreamTaskTitle);
    await downstreamDetails
      .getByRole("button", { name: new RegExp(upstreamTaskTitle) })
      .click();

    const upstreamDetails = page.getByRole("dialog", {
      name: upstreamTaskTitle,
    });
    await expect(upstreamDetails).toBeVisible();
  });

  test("user can find a task through global search and open its details", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Task Search Project ${crypto.randomUUID()}`;
    const taskTitle = `E2E Searchable Task ${crypto.randomUUID()}`;

    await createProject(page, projectName);

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    const createDialog = page.getByRole("dialog", { name: "Create task" });
    await createDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      "A task created to test global task search."
    );
    await selectOption(createDialog, page, "priority", "Low");
    await createDialog.getByRole("button", { name: "Create task" }).click();
    await expectToast(page, "Task created successfully");

    await page.goto("/projects");
    const searchInput = page.getByRole("searchbox", {
      name: "Search projects and tasks",
    });
    await searchInput.fill(taskTitle);
    await searchInput.press("Enter");

    await expect(page).toHaveURL(/\/search\?query=/);
    const taskResults = page
      .getByRole("heading", {
        name: "Tasks",
        exact: true,
      })
      .locator("..");
    await expect(
      taskResults.getByText(taskTitle, { exact: true })
    ).toBeVisible();
    await taskResults
      .getByRole("button", { name: `Open task ${taskTitle}`, exact: true })
      .click();

    const taskDetails = page.getByRole("dialog", { name: taskTitle });
    await expect(taskDetails).toBeVisible();
    await expect(
      taskDetails.getByText("A task created to test global task search.", {
        exact: true,
      })
    ).toBeVisible();
  });
});

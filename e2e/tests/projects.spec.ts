import crypto from "node:crypto";
import {
  test,
  expect,
  expectToast,
  loginAsUser,
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

function columnEditor(container: Locator, index: number) {
  return container
    .locator(`#column-${index}`)
    .locator(`xpath=ancestor::*[.//*[@id="column-description-${index}"]][1]`);
}

async function markColumnAsDone(container: Locator, index: number) {
  await columnEditor(container, index)
    .getByRole("button", { name: "Mark as done" })
    .click();
}

async function createProject(page: Page, name: string, description: string) {
  await page
    .getByRole("banner")
    .getByRole("button", { name: "Create project" })
    .click();

  const createDialog = page.getByRole("dialog", { name: "Create project" });
  await createDialog.locator("#name").fill(name);
  await fillRichText(createDialog.locator("#description"), page, description);
  await createDialog.getByRole("button", { name: "Create project" }).click();

  await expectToast(page, "Project created successfully");
}

function boardColumnHeadings(page: Page, names: string[]) {
  return page
    .getByRole("heading", { level: 3 })
    .filter({ hasText: new RegExp(`^(${names.join("|")})$`) });
}

function projectListLink(page: Page, name: string) {
  return page
    .getByRole("heading", { name: "Your Projects" })
    .locator("..")
    .getByRole("link", { name: new RegExp(name) });
}

test.describe("projects", () => {
  test("project owner can add and remove a member, granting and revoking project access", async ({
    authenticatedPage: ownerPage,
    backendURL,
    browser,
    request,
  }) => {
    const member = await registerUser(request, backendURL, {
      name: "E2E Project Member",
    });
    const memberPage = await loginAsUser(browser, member);
    const projectName = `E2E Membership ${crypto.randomUUID()}`;

    await createProject(
      ownerPage,
      projectName,
      "Created by the project membership e2e test"
    );
    await expect(projectListLink(memberPage, projectName)).toHaveCount(0);

    await projectListLink(ownerPage, projectName).click();
    await expect(
      ownerPage.getByRole("heading", { name: projectName })
    ).toBeVisible();

    await ownerPage.getByTitle("Add project member").click();
    const addMemberDialog = ownerPage.getByRole("dialog", {
      name: "Add project member",
    });
    await addMemberDialog.getByLabel("Email").fill(member.email);
    await addMemberDialog.getByRole("button", { name: "Add member" }).click();
    await expectToast(ownerPage, "Member added successfully");
    await expect(addMemberDialog).toHaveCount(0);

    await memberPage.reload();
    await projectListLink(memberPage, projectName).click();
    await expect(
      memberPage.getByRole("heading", { name: projectName })
    ).toBeVisible();
    await expect(
      boardColumnHeadings(memberPage, ["Pending", "Doing", "Done"])
    ).toHaveText(["Pending", "Doing", "Done"]);

    await ownerPage.getByTitle("View project members").click();
    let membersDialog = ownerPage.getByRole("dialog", {
      name: "Project members",
    });
    await expect(membersDialog.getByText("2 members")).toBeVisible();
    const memberRow = membersDialog.locator("article").filter({
      hasText: member.email,
    });
    await expect(memberRow.getByText("Member", { exact: true })).toBeVisible();
    await memberRow.hover();
    await memberRow
      .getByRole("button", { name: "Remove member from project" })
      .click();

    const removeMemberDialog = ownerPage.getByRole("dialog", {
      name: "Remove member from project",
    });
    await expect(
      removeMemberDialog.getByText(
        `Are you sure you want to remove ${member.name} from the project?`
      )
    ).toBeVisible();
    await removeMemberDialog.getByRole("button", { name: "Remove" }).click();
    await expect(removeMemberDialog).toHaveCount(0);

    membersDialog = ownerPage.getByRole("dialog", {
      name: "Project members",
    });
    await expect(membersDialog.getByText("1 member")).toBeVisible();
    await expect(membersDialog.getByText(member.email)).toHaveCount(0);

    await memberPage.goto("/projects");
    await expect(projectListLink(memberPage, projectName)).toHaveCount(0);

    await memberPage.context().close();
  });

  test("project creation shows validation errors for a missing name", async ({
    authenticatedPage: page,
  }) => {
    await page
      .getByRole("banner")
      .getByRole("button", { name: "Create project" })
      .click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });
    await createDialog.getByRole("button", { name: "Create project" }).click();

    await expect(createDialog.getByText("Name is required")).toBeVisible();
    await expect(createDialog).toBeVisible();
  });

  test("user can find a project through global search and open it", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Project Search ${crypto.randomUUID()}`;

    await createProject(
      page,
      projectName,
      "Created by the project search e2e test"
    );

    const searchInput = page.getByRole("searchbox", {
      name: "Search projects and tasks",
    });
    await searchInput.fill(projectName);
    await searchInput.press("Enter");

    await expect(page).toHaveURL(/\/search\?query=/);
    const projectResults = page
      .getByRole("heading", {
        name: "Projects",
        exact: true,
      })
      .locator("..");
    const projectResult = projectResults.getByRole("link", {
      name: new RegExp(projectName),
    });
    await expect(projectResult).toBeVisible();

    await projectResult.click();
    await expect(
      page.getByRole("heading", { name: projectName, exact: true })
    ).toBeVisible();
  });

  test("user can create a project with repository details and a custom workflow", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Workflow ${crypto.randomUUID()}`;
    const projectDescription =
      "Created with repository details and a custom board column";
    const repositoryURL =
      "https://github.com/example/project-chat-e2e-workflow";
    const repositoryOwner = "example";
    const repositoryName = "project-chat-e2e-workflow";
    const defaultBranch = "develop";
    const branchNamePrefix = "e2e/";
    const customColumnName = "Review";
    const customColumnDescription = "Validate work before it is done.";

    await page
      .getByRole("banner")
      .getByRole("button", { name: "Create project" })
      .click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });
    await createDialog.locator("#name").fill(projectName);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      projectDescription
    );
    await createDialog.locator("#repository_url").fill(repositoryURL);
    await createDialog.locator("#repository_owner").fill(repositoryOwner);
    await createDialog.locator("#repository_name").fill(repositoryName);
    await createDialog.locator("#default_branch").fill(defaultBranch);
    await createDialog.locator("#branch_name_prefix").fill(branchNamePrefix);

    await createDialog.getByRole("button", { name: "Add column" }).click();
    await createDialog.locator("#column-3").fill(customColumnName);
    await createDialog
      .locator("#column-description-3")
      .fill(customColumnDescription);
    await markColumnAsDone(createDialog, 3);

    await expect(
      createDialog.getByRole("button", { name: "Done column" })
    ).toHaveCount(1);
    await expect(
      columnEditor(createDialog, 3).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();
    await expect(
      columnEditor(createDialog, 2).getByRole("button", {
        name: "Mark as done",
      })
    ).toBeVisible();

    await createDialog.getByRole("button", { name: "Create project" }).click();

    await expectToast(page, "Project created successfully");
    await page.getByRole("link", { name: new RegExp(projectName) }).click();

    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();
    await expect(page.getByRole("heading", { name: "Pending" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Doing" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Done" })).toBeVisible();
    await expect(
      page.getByRole("heading", { name: customColumnName })
    ).toBeVisible();

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await expect(settingsDialog.locator("#repository_url")).toHaveValue(
      repositoryURL
    );
    await expect(settingsDialog.locator("#repository_owner")).toHaveValue(
      repositoryOwner
    );
    await expect(settingsDialog.locator("#repository_name")).toHaveValue(
      repositoryName
    );
    await expect(settingsDialog.locator("#default_branch")).toHaveValue(
      defaultBranch
    );
    await expect(settingsDialog.locator("#branch_name_prefix")).toHaveValue(
      branchNamePrefix
    );
    await expect(
      columnEditor(settingsDialog, 3).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();
    await expect(settingsDialog.locator("#column-description-3")).toHaveValue(
      customColumnDescription
    );
  });

  test("project creation only allows one done column", async ({
    authenticatedPage: page,
  }) => {
    await page
      .getByRole("banner")
      .getByRole("button", { name: "Create project" })
      .click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });

    await expect(
      createDialog.getByRole("button", { name: "Done column" })
    ).toHaveCount(1);
    await expect(
      columnEditor(createDialog, 2).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();

    await markColumnAsDone(createDialog, 0);

    await expect(
      createDialog.getByRole("button", { name: "Done column" })
    ).toHaveCount(1);
    await expect(
      columnEditor(createDialog, 0).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();
    await expect(
      columnEditor(createDialog, 2).getByRole("button", {
        name: "Mark as done",
      })
    ).toBeVisible();
  });

  test("user can update project columns from settings", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Column Settings ${crypto.randomUUID()}`;
    const projectDescription =
      "Created by the project column settings e2e test";
    const renamedColumnName = "Queued";
    const renamedColumnDescription = "New work starts here.";
    const addedColumnName = "QA";
    const addedColumnDescription = "Quality checks before completion.";

    await page
      .getByRole("banner")
      .getByRole("button", { name: "Create project" })
      .click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });
    await createDialog.locator("#name").fill(projectName);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      projectDescription
    );
    await createDialog.getByRole("button", { name: "Create project" }).click();

    await expectToast(page, "Project created successfully");
    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await settingsDialog.locator("#column-0").fill(renamedColumnName);
    await settingsDialog
      .locator("#column-description-0")
      .fill(renamedColumnDescription);
    await settingsDialog.getByRole("button", { name: "Add column" }).click();
    await settingsDialog.locator("#column-3").fill(addedColumnName);
    await settingsDialog
      .locator("#column-description-3")
      .fill(addedColumnDescription);
    await markColumnAsDone(settingsDialog, 3);

    await expect(
      settingsDialog.getByRole("button", { name: "Done column" })
    ).toHaveCount(1);
    await expect(
      columnEditor(settingsDialog, 3).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();
    await expect(
      columnEditor(settingsDialog, 2).getByRole("button", {
        name: "Mark as done",
      })
    ).toBeVisible();

    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project saved successfully");
    await expect(
      page.getByRole("heading", { name: renamedColumnName })
    ).toBeVisible();
    await expect(
      page.getByRole("heading", { name: addedColumnName })
    ).toBeVisible();

    await page.getByRole("button", { name: "Settings" }).click();

    const reopenedSettingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await expect(reopenedSettingsDialog.locator("#column-0")).toHaveValue(
      renamedColumnName
    );
    await expect(
      reopenedSettingsDialog.locator("#column-description-0")
    ).toHaveValue(renamedColumnDescription);
    await expect(reopenedSettingsDialog.locator("#column-3")).toHaveValue(
      addedColumnName
    );
    await expect(
      reopenedSettingsDialog.locator("#column-description-3")
    ).toHaveValue(addedColumnDescription);
    await expect(
      columnEditor(reopenedSettingsDialog, 3).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();
  });

  test("project owner can edit a column from the board and the changes persist", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Direct Column Edit ${crypto.randomUUID()}`;
    const updatedColumnName = "Ready";
    const updatedColumnDescription =
      "Work that is ready to count as completed.";
    const updatedColumnColor = "#7c3aed";

    await createProject(
      page,
      projectName,
      "Created by the direct column editing e2e test"
    );
    await projectListLink(page, projectName).click();
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();

    await page
      .getByRole("button", { name: "Open actions for Pending" })
      .click();
    await page.getByRole("menuitem", { name: "Edit column" }).click();

    let editDialog = page.getByRole("dialog", { name: "Edit column" });
    await expect(editDialog.getByLabel("Column name")).toHaveValue("Pending");
    await editDialog.getByLabel("Column name").fill(updatedColumnName);
    await editDialog
      .getByLabel("Column description")
      .fill(updatedColumnDescription);
    await editDialog.getByLabel("Color").fill(updatedColumnColor);
    await editDialog.getByRole("checkbox").check();
    await editDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Column updated successfully");
    await expect(editDialog).toHaveCount(0);
    await expect(
      page.getByRole("heading", { name: updatedColumnName })
    ).toBeVisible();
    await expect(page.getByRole("heading", { name: "Pending" })).toHaveCount(0);

    await page.reload();
    await expect(
      page.getByRole("heading", { name: updatedColumnName })
    ).toBeVisible();
    await page
      .getByRole("button", {
        name: `Open actions for ${updatedColumnName}`,
      })
      .click();
    await page.getByRole("menuitem", { name: "Edit column" }).click();

    editDialog = page.getByRole("dialog", { name: "Edit column" });
    await expect(editDialog.getByLabel("Column name")).toHaveValue(
      updatedColumnName
    );
    await expect(editDialog.getByLabel("Column description")).toHaveValue(
      updatedColumnDescription
    );
    await expect(editDialog.getByLabel("Color")).toHaveValue(
      updatedColumnColor
    );
    await expect(editDialog.getByRole("checkbox")).toBeChecked();
    await editDialog.getByRole("button", { name: "Cancel" }).click();

    await page.getByRole("button", { name: "Settings" }).click();
    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await expect(
      columnEditor(settingsDialog, 0).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();
    await expect(
      columnEditor(settingsDialog, 2).getByRole("button", {
        name: "Mark as done",
      })
    ).toBeVisible();
  });

  test("user can reorder project columns from settings", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Column Order ${crypto.randomUUID()}`;

    await createProject(
      page,
      projectName,
      "Created by the project column ordering e2e test"
    );

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();
    await expect(
      boardColumnHeadings(page, ["Pending", "Doing", "Done"])
    ).toHaveText(["Pending", "Doing", "Done"]);

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await settingsDialog.getByRole("button", { name: "Move Done up" }).click();
    await expect(settingsDialog.locator("#column-1")).toHaveValue("Done");
    await expect(settingsDialog.locator("#column-2")).toHaveValue("Doing");

    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project saved successfully");
    await expect(
      boardColumnHeadings(page, ["Pending", "Done", "Doing"])
    ).toHaveText(["Pending", "Done", "Doing"]);

    await page.getByRole("button", { name: "Settings" }).click();

    const reopenedSettingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await expect(reopenedSettingsDialog.locator("#column-0")).toHaveValue(
      "Pending"
    );
    await expect(reopenedSettingsDialog.locator("#column-1")).toHaveValue(
      "Done"
    );
    await expect(reopenedSettingsDialog.locator("#column-2")).toHaveValue(
      "Doing"
    );
  });

  test("user can delete project columns from settings", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Column Delete ${crypto.randomUUID()}`;

    await createProject(
      page,
      projectName,
      "Created by the project column deletion e2e test"
    );

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await settingsDialog.getByRole("button", { name: "Delete Doing" }).click();

    await expect(settingsDialog.getByText("Pending removals")).toBeVisible();
    await expect(settingsDialog.getByText("Doing")).toBeVisible();
    await expect(settingsDialog.locator("#column-0")).toHaveValue("Pending");
    await expect(settingsDialog.locator("#column-1")).toHaveValue("Done");
    await expect(settingsDialog.locator("#column-2")).toHaveCount(0);

    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project saved successfully");
    await expect(page.getByRole("heading", { name: "Pending" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Done" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Doing" })).toHaveCount(0);

    await page.getByRole("button", { name: "Settings" }).click();

    const reopenedSettingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await expect(reopenedSettingsDialog.locator("#column-0")).toHaveValue(
      "Pending"
    );
    await expect(reopenedSettingsDialog.locator("#column-1")).toHaveValue(
      "Done"
    );
    await expect(reopenedSettingsDialog.locator("#column-2")).toHaveCount(0);
    await expect(
      reopenedSettingsDialog.getByText("Pending removals")
    ).toHaveCount(0);
  });

  test("project owner can delete a populated column and reassign its tasks", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Column Reassignment ${crypto.randomUUID()}`;
    const taskTitle = `E2E Reassigned Task ${crypto.randomUUID()}`;

    await createProject(
      page,
      projectName,
      "Created by the populated column deletion e2e test"
    );

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();

    await page.getByRole("button", { name: "Open actions for Doing" }).click();
    await page.getByRole("menuitem", { name: "Add task" }).click();

    const createTaskDialog = page.getByRole("dialog", { name: "Create task" });
    await createTaskDialog.locator("#title").fill(taskTitle);
    await fillRichText(
      createTaskDialog.locator("#description"),
      page,
      "This task must survive removal of its workflow column."
    );
    await createTaskDialog.locator("#priority").click();
    await page.getByRole("option", { name: "Medium", exact: true }).click();
    await createTaskDialog.getByRole("button", { name: "Create task" }).click();

    await expectToast(page, "Task created successfully");
    await expect(page.getByText(taskTitle, { exact: true })).toBeVisible();

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await settingsDialog.getByRole("button", { name: "Delete Doing" }).click();
    await expect(settingsDialog.getByText("Pending removals")).toBeVisible();

    await settingsDialog.getByRole("combobox").click();
    await page.getByRole("option", { name: "Done", exact: true }).click();
    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project saved successfully");
    await expect(page.getByRole("heading", { name: "Doing" })).toHaveCount(0);
    await expect(page.getByText(taskTitle, { exact: true })).toBeVisible();

    await page.getByText(taskTitle, { exact: true }).click();
    const taskDetails = page.getByRole("dialog", { name: taskTitle });
    await expect(
      taskDetails
        .getByText("Status", { exact: true })
        .locator("..")
        .getByText("Done", { exact: true })
    ).toBeVisible();
  });

  test("user can create and update a project through the UI", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Project ${crypto.randomUUID()}`;
    const projectDescription = "Created by the projects e2e test";
    const updatedProjectName = `${projectName} updated`;
    const updatedProjectDescription =
      "Updated from project settings by the projects e2e test";

    await page
      .getByRole("banner")
      .getByRole("button", { name: "Create project" })
      .click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });
    await createDialog.locator("#name").fill(projectName);
    await fillRichText(
      createDialog.locator("#description"),
      page,
      projectDescription
    );
    await createDialog.getByRole("button", { name: "Create project" }).click();

    await expectToast(page, "Project created successfully");
    await expect(
      page.getByRole("link", { name: new RegExp(projectName) })
    ).toBeVisible();

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page).toHaveURL(/\/projects\/[^/]+$/);
    await expect(
      page.getByRole("heading", { name: projectName })
    ).toBeVisible();

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await expect(settingsDialog.locator("#name")).toHaveValue(projectName);
    await settingsDialog.locator("#name").fill(updatedProjectName);
    await fillRichText(
      settingsDialog.locator("#description"),
      page,
      updatedProjectDescription
    );
    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project saved successfully");
    await expect(
      page.getByRole("heading", { name: updatedProjectName })
    ).toBeVisible();

    await page.getByRole("button", { name: "View details" }).click();

    const detailsSheet = page.getByRole("dialog", { name: updatedProjectName });
    await expect(
      detailsSheet.getByText(updatedProjectDescription)
    ).toBeVisible();
  });
});

import crypto from "node:crypto";
import {
  addProjectMember,
  createProjectThroughUI,
  test,
  expect,
  expectToast,
  loginAsUser,
  openProjectMembersSettings,
} from "../src/fixtures/authenticated-page.js";
import { registerUser } from "../src/fixtures/test-user.js";
import { fillRichTextList } from "../src/fixtures/rich-text-editor.js";
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
    .locator(`#column-${index}, #column-name-${index}`)
    .first()
    .locator(`xpath=ancestor::*[.//*[@id="column-description-${index}"]][1]`);
}

async function markColumnAsDone(container: Locator, index: number) {
  const editor = columnEditor(container, index);
  const settingsControl = editor.getByRole("button", { name: /Done column/ });

  if ((await settingsControl.count()) > 0) {
    await settingsControl.click();
    return;
  }

  await editor.getByRole("button", { name: "Mark as done" }).click();
}

async function openSettingsColumn(container: Locator, name: string) {
  await container
    .locator('[data-slot="accordion-trigger"]')
    .filter({ hasText: new RegExp(`^${name}`) })
    .click();
}

function boardColumnHeadings(page: Page, names: string[]) {
  return page
    .getByRole("heading", { level: 3 })
    .filter({ hasText: new RegExp(`^(${names.join("|")})$`) });
}

function projectListLink(page: Page, name: string) {
  return page
    .getByRole("region", { name: "Projects" })
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

    await createProjectThroughUI(ownerPage, projectName, {
      description: "Created by the project membership e2e test",
    });
    await expect(projectListLink(memberPage, projectName)).toHaveCount(0);

    await projectListLink(ownerPage, projectName).click();
    await expect(
      ownerPage.getByRole("button", { name: projectName })
    ).toBeVisible();

    await addProjectMember(ownerPage, member.email);

    await memberPage.reload();
    await projectListLink(memberPage, projectName).click();
    await expect(
      memberPage.getByRole("button", { name: projectName })
    ).toBeVisible();
    await expect(
      boardColumnHeadings(memberPage, ["Pending", "Doing", "Done"])
    ).toHaveText(["Pending", "Doing", "Done"]);

    await openProjectMembersSettings(ownerPage);
    await expect(ownerPage.getByText("MEMBERS • 2")).toBeVisible();
    const memberRow = ownerPage.locator("article").filter({
      hasText: member.email,
    });
    await expect(memberRow.getByText("member", { exact: true })).toBeVisible();
    await memberRow
      .getByRole("button", { name: `Remove ${member.name} from project` })
      .click();

    const removeMemberDialog = ownerPage.getByRole("dialog", {
      name: "Remove member from project?",
    });
    await expect(
      removeMemberDialog.getByText(
        `Are you sure you want to remove ${member.name} from the project?`
      )
    ).toBeVisible();
    await removeMemberDialog
      .getByRole("button", { name: "Remove member" })
      .click();
    await expect(removeMemberDialog).toHaveCount(0);

    await expect(ownerPage.getByText("MEMBERS • 1")).toBeVisible();
    await expect(ownerPage.getByText(member.email)).toHaveCount(0);

    await memberPage.goto("/projects");
    await expect(projectListLink(memberPage, projectName)).toHaveCount(0);

    await memberPage.context().close();
  });

  test("project creation shows validation errors for a missing name", async ({
    authenticatedPage: page,
  }) => {
    await page.getByRole("button", { name: "New project" }).first().click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });
    await createDialog.getByRole("button", { name: "Create project" }).click();

    await expect(createDialog.getByText("Name is required")).toBeVisible();
    await expect(createDialog).toBeVisible();
  });

  test("project descriptions preserve numbered lists through settings and reload", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Project List ${crypto.randomUUID()}`;
    const listItems = ["Create the project", "Review the workflow"];

    await page.getByRole("button", { name: "New project" }).first().click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });
    await createDialog.locator("#name").fill(projectName);
    await fillRichTextList(createDialog, "numbered", listItems);
    await createDialog.getByRole("button", { name: "Create project" }).click();

    await expectToast(page, "Project created successfully");
    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await page.getByRole("link", { name: "Settings" }).click();

    const settingsForm = page.locator("#general-project-settings");
    await expect(settingsForm.locator("#description ol > li")).toHaveText(
      listItems
    );

    await page.reload();
    await expect(settingsForm.locator("#description ol > li")).toHaveText(
      listItems
    );
  });

  test("user can find a project through global search and open it", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Project Search ${crypto.randomUUID()}`;

    await createProjectThroughUI(page, projectName, {
      description: "Created by the project search e2e test",
    });

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
      page.getByRole("button", { name: projectName, exact: true })
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

    await createProjectThroughUI(page, projectName, {
      description: projectDescription,
      repository: {
        url: repositoryURL,
        owner: repositoryOwner,
        name: repositoryName,
        defaultBranch,
        branchNamePrefix,
      },
      columns: [
        { name: "Pending" },
        { name: "Doing" },
        { name: "Done" },
        {
          name: customColumnName,
          description: customColumnDescription,
          isDone: true,
        },
      ],
    });
    await page.getByRole("link", { name: new RegExp(projectName) }).click();

    await expect(page.getByRole("button", { name: projectName })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Pending" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Doing" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Done" })).toBeVisible();
    await expect(
      page.getByRole("heading", { name: customColumnName })
    ).toBeVisible();

    await page.getByRole("link", { name: "Settings" }).click();

    const generalSettings = page.locator("#general-project-settings");
    await expect(generalSettings.locator("#repository_url")).toHaveValue(
      repositoryURL
    );
    await expect(generalSettings.locator("#repository_owner")).toHaveValue(
      repositoryOwner
    );
    await expect(generalSettings.locator("#repository_name")).toHaveValue(
      repositoryName
    );
    await expect(generalSettings.locator("#default_branch")).toHaveValue(
      defaultBranch
    );
    await expect(generalSettings.locator("#branch_name_prefix")).toHaveValue(
      branchNamePrefix
    );

    await page.getByRole("link", { name: "Columns", exact: true }).click();
    const columnSettings = page.locator("#columns-project-settings");
    await columnSettings.getByText(customColumnName, { exact: true }).click();
    await expect(columnSettings.locator("#column-name-3")).toHaveValue(
      customColumnName
    );
    await expect(columnSettings.locator("#column-description-3")).toHaveValue(
      customColumnDescription
    );
    await expect(
      columnEditor(columnSettings, 3).getByRole("button", {
        name: /Done column/,
      })
    ).toHaveAttribute("aria-pressed", "true");
  });

  test("project creation only allows one done column", async ({
    authenticatedPage: page,
  }) => {
    await page.getByRole("button", { name: "New project" }).first().click();

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

    await createProjectThroughUI(page, projectName, {
      description: projectDescription,
    });
    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page.getByRole("button", { name: projectName })).toBeVisible();

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();

    const settingsDialog = page.locator("#columns-project-settings");
    await settingsDialog.locator("#column-name-0").fill(renamedColumnName);
    await settingsDialog
      .locator("#column-description-0")
      .fill(renamedColumnDescription);
    await settingsDialog.getByRole("button", { name: "Add column" }).click();
    await settingsDialog.locator("#column-name-3").fill(addedColumnName);
    await settingsDialog
      .locator("#column-description-3")
      .fill(addedColumnDescription);
    await markColumnAsDone(settingsDialog, 3);

    await expect(
      columnEditor(settingsDialog, 3).getByRole("button", {
        name: /Done column/,
      })
    ).toHaveAttribute("aria-pressed", "true");
    await openSettingsColumn(settingsDialog, "Done");
    await expect(
      columnEditor(settingsDialog, 2).getByRole("button", {
        name: /Done column/,
      })
    ).toHaveAttribute("aria-pressed", "false");

    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project columns saved successfully");
    await page.getByRole("link", { name: "Back to board" }).click();
    await expect(
      page.getByRole("heading", { name: renamedColumnName })
    ).toBeVisible();
    await expect(
      page.getByRole("heading", { name: addedColumnName })
    ).toBeVisible();

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();

    const reopenedSettingsDialog = page.locator("#columns-project-settings");
    await expect(reopenedSettingsDialog.locator("#column-name-0")).toHaveValue(
      renamedColumnName
    );
    await expect(
      reopenedSettingsDialog.locator("#column-description-0")
    ).toHaveValue(renamedColumnDescription);
    await reopenedSettingsDialog
      .getByText(addedColumnName, { exact: true })
      .click();
    await expect(reopenedSettingsDialog.locator("#column-name-3")).toHaveValue(
      addedColumnName
    );
    await expect(
      reopenedSettingsDialog.locator("#column-description-3")
    ).toHaveValue(addedColumnDescription);
    await expect(
      columnEditor(reopenedSettingsDialog, 3).getByRole("button", {
        name: /Done column/,
      })
    ).toHaveAttribute("aria-pressed", "true");
  });

  test("project owner can edit a column from the board and the changes persist", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Direct Column Edit ${crypto.randomUUID()}`;
    const updatedColumnName = "Ready";
    const updatedColumnDescription =
      "Work that is ready to count as completed.";
    const updatedColumnColor = "#7c3aed";

    await createProjectThroughUI(page, projectName, {
      description: "Created by the direct column editing e2e test",
    });
    await projectListLink(page, projectName).click();
    await expect(page.getByRole("button", { name: projectName })).toBeVisible();

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

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();
    const settingsDialog = page.locator("#columns-project-settings");
    await expect(
      columnEditor(settingsDialog, 0).getByRole("button", {
        name: /Done column/,
      })
    ).toHaveAttribute("aria-pressed", "true");
    await openSettingsColumn(settingsDialog, "Done");
    await expect(
      columnEditor(settingsDialog, 2).getByRole("button", {
        name: /Done column/,
      })
    ).toHaveAttribute("aria-pressed", "false");
  });

  test("user can reorder project columns from settings", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Column Order ${crypto.randomUUID()}`;

    await createProjectThroughUI(page, projectName, {
      description: "Created by the project column ordering e2e test",
    });

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page.getByRole("button", { name: projectName })).toBeVisible();
    await expect(
      boardColumnHeadings(page, ["Pending", "Doing", "Done"])
    ).toHaveText(["Pending", "Doing", "Done"]);

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();

    const settingsDialog = page.locator("#columns-project-settings");
    await settingsDialog.getByRole("button", { name: "Move Done up" }).click();
    await expect(
      settingsDialog.locator('[data-slot="accordion-trigger"]')
    ).toHaveText([/Pending/, /Done/, /Doing/]);

    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project columns saved successfully");
    await page.getByRole("link", { name: "Back to board" }).click();
    await expect(
      boardColumnHeadings(page, ["Pending", "Done", "Doing"])
    ).toHaveText(["Pending", "Done", "Doing"]);

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();

    const reopenedSettingsDialog = page.locator("#columns-project-settings");
    await expect(
      reopenedSettingsDialog.locator('[data-slot="accordion-trigger"]')
    ).toHaveText([/Pending/, /Done/, /Doing/]);
  });

  test("user can delete project columns from settings", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Column Delete ${crypto.randomUUID()}`;

    await createProjectThroughUI(page, projectName, {
      description: "Created by the project column deletion e2e test",
    });

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page.getByRole("button", { name: projectName })).toBeVisible();

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();

    const settingsDialog = page.locator("#columns-project-settings");
    await openSettingsColumn(settingsDialog, "Doing");
    await settingsDialog.getByRole("button", { name: "Delete Doing" }).click();

    await expect(settingsDialog.getByText(/Pending removals/i)).toBeVisible();
    await expect(settingsDialog.getByText("Doing")).toBeVisible();
    await expect(
      settingsDialog.locator('[data-slot="accordion-trigger"]')
    ).toHaveText([/Pending/, /Done/]);

    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project columns saved successfully");
    await page.getByRole("link", { name: "Back to board" }).click();
    await expect(page.getByRole("heading", { name: "Pending" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Done" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Doing" })).toHaveCount(0);

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();

    const reopenedSettingsDialog = page.locator("#columns-project-settings");
    await expect(
      reopenedSettingsDialog.locator('[data-slot="accordion-trigger"]')
    ).toHaveText([/Pending/, /Done/]);
    await expect(
      reopenedSettingsDialog.getByText(/Pending removals/i)
    ).toHaveCount(0);
  });

  test("project owner can delete a populated column and reassign its tasks", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Column Reassignment ${crypto.randomUUID()}`;
    const taskTitle = `E2E Reassigned Task ${crypto.randomUUID()}`;

    await createProjectThroughUI(page, projectName, {
      description: "Created by the populated column deletion e2e test",
    });

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page.getByRole("button", { name: projectName })).toBeVisible();

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

    await page.getByRole("link", { name: "Settings" }).click();
    await page.getByRole("link", { name: "Columns", exact: true }).click();

    const settingsDialog = page.locator("#columns-project-settings");
    await openSettingsColumn(settingsDialog, "Doing");
    await settingsDialog.getByRole("button", { name: "Delete Doing" }).click();
    await expect(settingsDialog.getByText(/Pending removals/i)).toBeVisible();

    await settingsDialog.getByRole("combobox").click();
    await page.getByRole("option", { name: "Done", exact: true }).click();
    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project columns saved successfully");
    await page.getByRole("link", { name: "Back to board" }).click();
    await expect(page.getByRole("heading", { name: "Doing" })).toHaveCount(0);
    await expect(page.getByText(taskTitle, { exact: true })).toBeVisible();

    await page.getByText(taskTitle, { exact: true }).click();
    const taskDetails = page.getByRole("dialog", { name: taskTitle });
    await expect(
      taskDetails
        .getByText("Column", { exact: true })
        .locator("..")
        .getByText("Done", { exact: true })
    ).toBeVisible();
  });

  test("user can create and update project details through the UI", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Project ${crypto.randomUUID()}`;
    const projectDescription = "Created by the projects e2e test";
    const updatedProjectName = `${projectName} updated`;
    const updatedProjectDescription =
      "Updated from project settings by the projects e2e test";
    const repositoryURL =
      "https://github.com/example/project-chat-e2e-settings";
    const repositoryOwner = "example";
    const repositoryName = "project-chat-e2e-settings";
    const defaultBranch = "release";
    const branchNamePrefix = "e2e/settings/";

    await createProjectThroughUI(page, projectName, {
      description: projectDescription,
    });
    await expect(
      page.getByRole("link", { name: new RegExp(projectName) })
    ).toBeVisible();

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page).toHaveURL(/\/projects\/[^/]+$/);
    await expect(page.getByRole("button", { name: projectName })).toBeVisible();

    await page.getByRole("link", { name: "Settings" }).click();

    const settingsDialog = page.locator("#general-project-settings");
    await expect(settingsDialog.locator("#name")).toHaveValue(projectName);
    await settingsDialog.locator("#name").fill(updatedProjectName);
    await fillRichText(
      settingsDialog.locator("#description"),
      page,
      updatedProjectDescription
    );
    await settingsDialog.locator("#repository_url").fill(repositoryURL);
    await settingsDialog.locator("#repository_owner").fill(repositoryOwner);
    await settingsDialog.locator("#repository_name").fill(repositoryName);
    await settingsDialog.locator("#default_branch").fill(defaultBranch);
    await settingsDialog.locator("#branch_name_prefix").fill(branchNamePrefix);
    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project saved successfully");
    await expect(
      page.getByRole("link", { name: updatedProjectName, exact: true })
    ).toBeVisible();

    await page.reload();
    await expect(
      page.getByRole("link", { name: updatedProjectName, exact: true })
    ).toBeVisible();

    const reopenedSettingsDialog = page.locator("#general-project-settings");
    await expect(reopenedSettingsDialog.locator("#name")).toHaveValue(
      updatedProjectName
    );
    await expect(reopenedSettingsDialog.locator("#description")).toHaveText(
      updatedProjectDescription
    );
    await expect(reopenedSettingsDialog.locator("#repository_url")).toHaveValue(
      repositoryURL
    );
    await expect(
      reopenedSettingsDialog.locator("#repository_owner")
    ).toHaveValue(repositoryOwner);
    await expect(
      reopenedSettingsDialog.locator("#repository_name")
    ).toHaveValue(repositoryName);
    await expect(reopenedSettingsDialog.locator("#default_branch")).toHaveValue(
      defaultBranch
    );
    await expect(
      reopenedSettingsDialog.locator("#branch_name_prefix")
    ).toHaveValue(branchNamePrefix);
  });

  test("project owner can delete a project from the danger zone", async ({
    authenticatedPage: page,
  }) => {
    const projectName = `E2E Project Delete ${crypto.randomUUID()}`;

    await createProjectThroughUI(page, projectName, {
      description: "Created by the project deletion e2e test",
    });

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page.getByRole("button", { name: projectName })).toBeVisible();

    await page.getByRole("link", { name: "Settings" }).click();
    await expect(page.locator("#general-project-settings")).toBeVisible();
    await page.getByRole("button", { name: "Delete" }).click();

    const confirmDialog = page.getByRole("dialog", {
      name: `Delete ${projectName}?`,
    });
    await expect(confirmDialog).toBeVisible();
    await confirmDialog.getByRole("button", { name: "Delete project" }).click();

    await expectToast(page, "Project deleted successfully");
    await expect(page).toHaveURL(/\/projects\/?$/);
    await expect(
      page.getByRole("link", { name: new RegExp(projectName) })
    ).toHaveCount(0);
  });
});

import crypto from "node:crypto";
import {
  test,
  expect,
  expectToast,
} from "../src/fixtures/authenticated-page.js";
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
    .locator(
      `xpath=ancestor::*[.//*[@id="column-description-${index}"]][1]`
    );
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
  await createDialog
    .getByRole("button", { name: "Create project" })
    .click();

  await expectToast(page, "Project created successfully");
}

function boardColumnHeadings(page: Page, names: string[]) {
  return page
    .getByRole("heading", { level: 3 })
    .filter({ hasText: new RegExp(`^(${names.join("|")})$`) });
}

test.describe("projects", () => {
  test("project creation shows validation errors for a missing name", async ({
    authenticatedPage: page,
  }) => {
    await page
      .getByRole("banner")
      .getByRole("button", { name: "Create project" })
      .click();

    const createDialog = page.getByRole("dialog", { name: "Create project" });
    await createDialog
      .getByRole("button", { name: "Create project" })
      .click();

    await expect(createDialog.getByText("Name is required")).toBeVisible();
    await expect(createDialog).toBeVisible();
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

    await expect(createDialog.getByRole("button", { name: "Done column" }))
      .toHaveCount(1);
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

    await createDialog
      .getByRole("button", { name: "Create project" })
      .click();

    await expectToast(page, "Project created successfully");
    await page.getByRole("link", { name: new RegExp(projectName) }).click();

    await expect(page.getByRole("heading", { name: projectName })).toBeVisible();
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

    await expect(createDialog.getByRole("button", { name: "Done column" }))
      .toHaveCount(1);
    await expect(
      columnEditor(createDialog, 2).getByRole("button", {
        name: "Done column",
      })
    ).toBeVisible();

    await markColumnAsDone(createDialog, 0);

    await expect(createDialog.getByRole("button", { name: "Done column" }))
      .toHaveCount(1);
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
    await createDialog
      .getByRole("button", { name: "Create project" })
      .click();

    await expectToast(page, "Project created successfully");
    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page.getByRole("heading", { name: projectName })).toBeVisible();

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

    await expect(settingsDialog.getByRole("button", { name: "Done column" }))
      .toHaveCount(1);
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
    await expect(page.getByRole("heading", { name: projectName })).toBeVisible();
    await expect(boardColumnHeadings(page, ["Pending", "Doing", "Done"]))
      .toHaveText(["Pending", "Doing", "Done"]);

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await settingsDialog.getByRole("button", { name: "Move Done up" }).click();
    await expect(settingsDialog.locator("#column-1")).toHaveValue("Done");
    await expect(settingsDialog.locator("#column-2")).toHaveValue("Doing");

    await settingsDialog.getByRole("button", { name: "Save changes" }).click();

    await expectToast(page, "Project saved successfully");
    await expect(boardColumnHeadings(page, ["Pending", "Done", "Doing"]))
      .toHaveText(["Pending", "Done", "Doing"]);

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
    await expect(page.getByRole("heading", { name: projectName })).toBeVisible();

    await page.getByRole("button", { name: "Settings" }).click();

    const settingsDialog = page.getByRole("dialog", {
      name: "Project settings",
    });
    await settingsDialog
      .getByRole("button", { name: "Delete Doing" })
      .click();

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
    await createDialog
      .getByRole("button", { name: "Create project" })
      .click();

    await expectToast(page, "Project created successfully");
    await expect(
      page.getByRole("link", { name: new RegExp(projectName) })
    ).toBeVisible();

    await page.getByRole("link", { name: new RegExp(projectName) }).click();
    await expect(page).toHaveURL(/\/projects\/[^/]+$/);
    await expect(page.getByRole("heading", { name: projectName })).toBeVisible();

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

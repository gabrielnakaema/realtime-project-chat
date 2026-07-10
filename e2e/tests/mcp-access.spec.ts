import crypto from "node:crypto";
import {
  expect,
  expectToast,
  test,
} from "../src/fixtures/authenticated-page.js";
import type { Page } from "@playwright/test";

async function createMCPKey(page: Page, keyName: string) {
  await page.getByRole("button", { name: "Create key" }).first().click();
  const createDialog = page.getByRole("dialog", { name: "Create MCP key" });
  await createDialog.locator("#mcp-key-name").fill(keyName);
  await createDialog
    .getByText("Read tasks", { exact: true })
    .locator("xpath=ancestor::label[1]")
    .getByRole("checkbox")
    .check();
  await createDialog.getByRole("button", { name: "Create key" }).click();

  return page.getByRole("dialog", { name: "Save your new MCP key" });
}

test("user can create an MCP API key and save its one-time secret", async ({
  authenticatedPage: page,
}) => {
  const keyName = `E2E MCP Key ${crypto.randomUUID()}`;

  await page.goto("/mcp-access");
  await expect(
    page.getByRole("heading", { name: "MCP Access", exact: true })
  ).toBeVisible();

  const revealDialog = await createMCPKey(page, keyName);
  await expect(
    revealDialog.getByText("Shown once", { exact: true })
  ).toBeVisible();
  await expect(revealDialog.getByText(/^mcp_[a-f0-9]+_/)).toBeVisible();
  await revealDialog.getByRole("button", { name: "Copy key" }).click();
  await expectToast(page, "Secret copied");
  await revealDialog
    .getByRole("button", { name: "Close", exact: true })
    .first()
    .click();
  await expect(revealDialog).toHaveCount(0);

  await page.reload();
  const keyCard = page.locator("article").filter({ hasText: keyName });
  await expect(keyCard).toBeVisible();
  await expect(keyCard.getByText("Active", { exact: true })).toBeVisible();
  await expect(keyCard.getByText("Read tasks", { exact: true })).toBeVisible();
});

test("user can revoke an MCP API key and the revocation persists", async ({
  authenticatedPage: page,
}) => {
  const keyName = `E2E Revoke MCP Key ${crypto.randomUUID()}`;

  await page.goto("/mcp-access");
  const revealDialog = await createMCPKey(page, keyName);
  await revealDialog
    .getByRole("button", { name: "Close", exact: true })
    .first()
    .click();
  const dismissDialog = page.getByRole("dialog", {
    name: "Leave without copying?",
  });
  await expect(dismissDialog).toBeVisible();
  await dismissDialog.getByRole("button", { name: "Close anyway" }).click();

  const keyCard = page.locator("article").filter({ hasText: keyName });
  await expect(keyCard.getByText("Active", { exact: true })).toBeVisible();
  await keyCard.getByRole("button", { name: "Revoke", exact: true }).click();

  const revokeDialog = page.getByRole("dialog", { name: "Revoke MCP key" });
  await expect(revokeDialog.getByText(keyName, { exact: true })).toBeVisible();
  await revokeDialog.getByRole("button", { name: "Revoke key" }).click();
  await expectToast(page, "MCP key revoked");
  await expect(keyCard.getByText("Revoked", { exact: true })).toBeVisible();
  await expect(
    keyCard.getByRole("button", { name: "Revoke", exact: true })
  ).toHaveCount(0);

  await page.reload();
  await expect(
    page
      .locator("article")
      .filter({ hasText: keyName })
      .getByText("Revoked", { exact: true })
  ).toBeVisible();
});

test("user can update an MCP API key's name and permissions", async ({
  authenticatedPage: page,
}) => {
  const keyName = `E2E Edit MCP Key ${crypto.randomUUID()}`;
  const updatedKeyName = `${keyName} updated`;

  await page.goto("/mcp-access");
  const revealDialog = await createMCPKey(page, keyName);
  await revealDialog.getByRole("button", { name: "Copy key" }).click();
  await expectToast(page, "Secret copied");
  await revealDialog
    .getByRole("button", { name: "Close", exact: true })
    .first()
    .click();

  const keyCard = page.locator("article").filter({ hasText: keyName });
  await keyCard.getByRole("button", { name: "Edit", exact: true }).click();

  const editDialog = page.getByRole("dialog", { name: "Edit MCP key" });
  await editDialog.locator("#mcp-key-name").fill(updatedKeyName);
  await editDialog
    .getByText("Read tasks", { exact: true })
    .locator("xpath=ancestor::label[1]")
    .getByRole("checkbox")
    .uncheck();
  await editDialog
    .getByText("Create tasks", { exact: true })
    .locator("xpath=ancestor::label[1]")
    .getByRole("checkbox")
    .check();
  await editDialog.getByRole("button", { name: "Save changes" }).click();
  await expectToast(page, "Key updated");

  const updatedKeyCard = page
    .locator("article")
    .filter({ hasText: updatedKeyName });
  await expect(
    updatedKeyCard.getByText("Create tasks", { exact: true })
  ).toBeVisible();
  await expect(
    updatedKeyCard.getByText("Read tasks", { exact: true })
  ).toHaveCount(0);

  await page.reload();
  const persistedKeyCard = page
    .locator("article")
    .filter({ hasText: updatedKeyName });
  await expect(
    persistedKeyCard.getByText("Create tasks", { exact: true })
  ).toBeVisible();
  await expect(
    persistedKeyCard.getByText("Read tasks", { exact: true })
  ).toHaveCount(0);
});

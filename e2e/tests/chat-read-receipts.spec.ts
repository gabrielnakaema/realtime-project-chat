import crypto from "node:crypto";
import { test, expect } from "@playwright/test";
import { registerUser } from "../src/fixtures/test-user.js";
import {
  backendURL,
  expectToast,
  loginAsUser,
} from "../src/fixtures/authenticated-page.js";

test("project chat sender can see when a collaborator reads a message", async ({
  browser,
  request,
}) => {
  const sender = await registerUser(request, backendURL, {
    name: "Read Receipt Sender",
  });
  const recipient = await registerUser(request, backendURL, {
    name: "Read Receipt Recipient",
  });
  const senderPage = await loginAsUser(browser, sender);
  const recipientPage = await loginAsUser(browser, recipient);
  const projectName = `E2E Read Receipts Project ${crypto.randomUUID()}`;
  const messageContent = `E2E read receipt message ${crypto.randomUUID()}`;

  await senderPage
    .getByRole("banner")
    .getByRole("button", { name: "Create project" })
    .click();
  const createProjectDialog = senderPage.getByRole("dialog", {
    name: "Create project",
  });
  await createProjectDialog.locator("#name").fill(projectName);
  await createProjectDialog
    .locator("#description")
    .pressSequentially("Created by the chat read-receipts e2e test.");
  await createProjectDialog
    .getByRole("button", { name: "Create project" })
    .click();
  await expectToast(senderPage, "Project created successfully");

  await senderPage.getByRole("link", { name: projectName }).click();
  const projectId = senderPage.url().split("/projects/")[1];
  await senderPage.getByTitle("Add project member").click();
  const addMemberDialog = senderPage.getByRole("dialog", {
    name: "Add project member",
  });
  await addMemberDialog.getByLabel("Email").fill(recipient.email);
  await addMemberDialog.getByRole("button", { name: "Add member" }).click();
  await expectToast(senderPage, "Member added successfully");

  await senderPage.getByRole("link", { name: "Chat" }).click();
  await recipientPage.goto(`/projects/${projectId}/chat`);
  await expect(senderPage.locator('form button[type="submit"]')).toBeEnabled({
    timeout: 15_000,
  });
  await expect(recipientPage.locator('form button[type="submit"]')).toBeEnabled(
    {
      timeout: 15_000,
    }
  );
  await expect(
    senderPage
      .getByRole("button", { name: "R", exact: true })
      .locator("div.border-green-500")
      .first()
  ).toBeVisible({ timeout: 15_000 });

  const recipientReadResponsePromise = recipientPage.waitForResponse(
    (response) => {
      const url = new URL(response.url());

      return (
        response.request().method() === "POST" &&
        /\/chats\/[^/]+\/read$/.test(url.pathname)
      );
    }
  );
  await senderPage.getByPlaceholder("Message...").fill(messageContent);
  await senderPage.getByPlaceholder("Message...").press("Enter");

  await expect(recipientPage.getByText(messageContent)).toBeVisible({
    timeout: 15_000,
  });
  expect((await recipientReadResponsePromise).ok()).toBe(true);

  const senderMessage = senderPage
    .getByText(messageContent, { exact: true })
    .locator("xpath=ancestor::div[contains(@class, 'group flex gap-2')][1]");
  await senderMessage.hover();
  await senderMessage.getByRole("button", { name: "Message details" }).click();

  const readStatusSheet = senderPage.getByRole("dialog", {
    name: "Read status",
  });
  await expect(readStatusSheet).toBeVisible();
  await expect(
    readStatusSheet.getByText(recipient.name, { exact: true })
  ).toBeVisible({ timeout: 15_000 });

  await senderPage.context().close();
  await recipientPage.context().close();
});

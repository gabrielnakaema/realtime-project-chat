import crypto from "node:crypto";
import { test, expect } from "@playwright/test";
import { registerUser } from "../src/fixtures/test-user.js";
import { backendURL, loginAsUser } from "../src/fixtures/authenticated-page.js";

test("a chat message sent by one member appears for another member in real time", async ({
  browser,
  request,
}) => {
  const userA = await registerUser(request, backendURL, { name: "Alice E2E" });
  const userB = await registerUser(request, backendURL, { name: "Bob E2E" });

  const pageA = await loginAsUser(browser, userA);
  const pageB = await loginAsUser(browser, userB);

  const projectName = `E2E Project ${crypto.randomUUID()}`;

  await pageA
    .getByRole("banner")
    .getByRole("button", { name: "Create project" })
    .click();
  await pageA.locator("#name").fill(projectName);
  await pageA.locator("#description").click();
  await pageA
    .locator("#description")
    .pressSequentially("Created by the chat-realtime e2e test");
  await pageA
    .getByRole("dialog")
    .getByRole("button", { name: "Create project" })
    .click();
  await expect(pageA.getByText("Project created successfully")).toBeVisible();

  await pageA.getByRole("link", { name: projectName }).click();
  await expect(pageA).toHaveURL(/\/projects\/[^/]+$/);
  const projectId = pageA.url().split("/projects/")[1];

  await pageA.getByTitle("Add project member").click();
  await pageA.locator("#email").fill(userB.email);
  await pageA.getByRole("button", { name: "Add member" }).click();
  await expect(pageA.getByText("Member added successfully")).toBeVisible();

  await pageA.getByRole("link", { name: "Chat" }).click();
  await pageB.goto(`/projects/${projectId}/chat`);

  const submitButton = pageA.locator('form button[type="submit"]');
  await expect(submitButton).toBeEnabled({ timeout: 15_000 });

  const messageContent = `Hello from Alice ${crypto.randomUUID()}`;
  await pageA.getByPlaceholder("Message...").fill(messageContent);
  await pageA.getByPlaceholder("Message...").press("Enter");

  await expect(pageB.getByText(messageContent)).toBeVisible({ timeout: 5_000 });

  await pageA.context().close();
  await pageB.context().close();
});

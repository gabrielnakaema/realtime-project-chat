import crypto from "node:crypto";
import { test, expect } from "@playwright/test";
import { registerUser } from "../src/fixtures/test-user.js";
import {
  backendURL,
  expectToast,
  loginAsUser,
} from "../src/fixtures/authenticated-page.js";

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
  await expectToast(pageA, "Project created successfully");

  await pageA.getByRole("link", { name: projectName }).click();
  await expect(pageA).toHaveURL(/\/projects\/[^/]+$/);
  const projectId = pageA.url().split("/projects/")[1];

  await pageA.getByTitle("Add project member").click();
  await pageA.locator("#email").fill(userB.email);
  await pageA.getByRole("button", { name: "Add member" }).click();
  await expectToast(pageA, "Member added successfully");

  await pageA.getByRole("link", { name: "Chat" }).click();
  await pageB.goto(`/projects/${projectId}/chat`);

  await expect(pageA.locator('form button[type="submit"]')).toBeEnabled({
    timeout: 15_000,
  });
  await expect(pageB.locator('form button[type="submit"]')).toBeEnabled({
    timeout: 15_000,
  });
  await expect(
    pageB
      .getByRole("button", { name: "B", exact: true })
      .locator("div.border-green-500")
  ).toBeVisible({ timeout: 15_000 });

  const messageContent = `Hello from Alice ${crypto.randomUUID()}`;
  await pageA.getByPlaceholder("Message...").fill(messageContent);

  const createMessageResponsePromise = pageA.waitForResponse((response) => {
    const url = new URL(response.url());

    return (
      response.request().method() === "POST" &&
      url.pathname.endsWith("/chats/messages")
    );
  });

  await pageA.getByPlaceholder("Message...").press("Enter");
  const createMessageResponse = await createMessageResponsePromise;
  expect(createMessageResponse.ok()).toBe(true);

  await expect(pageB.getByText(messageContent)).toBeVisible({
    timeout: 15_000,
  });

  await pageA.context().close();
  await pageB.context().close();
});

test("project collaborators can start a direct message and receive it in real time", async ({
  browser,
  request,
}) => {
  const userA = await registerUser(request, backendURL, {
    name: "Alice Direct E2E",
  });
  const userB = await registerUser(request, backendURL, {
    name: "Bob Direct E2E",
  });

  const pageA = await loginAsUser(browser, userA);
  const pageB = await loginAsUser(browser, userB);

  const projectName = `E2E Direct Message Project ${crypto.randomUUID()}`;
  await pageA
    .getByRole("banner")
    .getByRole("button", { name: "Create project" })
    .click();
  const createProjectDialog = pageA.getByRole("dialog", {
    name: "Create project",
  });
  await createProjectDialog.locator("#name").fill(projectName);
  await createProjectDialog
    .locator("#description")
    .pressSequentially("Created by the direct-message e2e test");
  await createProjectDialog
    .getByRole("button", { name: "Create project" })
    .click();
  await expectToast(pageA, "Project created successfully");

  await pageA.getByRole("link", { name: projectName }).click();
  await pageA.getByTitle("Add project member").click();
  const addMemberDialog = pageA.getByRole("dialog", {
    name: "Add project member",
  });
  await addMemberDialog.getByLabel("Email").fill(userB.email);
  await addMemberDialog.getByRole("button", { name: "Add member" }).click();
  await expectToast(pageA, "Member added successfully");

  await pageA.goto("/projects");
  await pageA.getByRole("button", { name: "Messages" }).click();
  await pageA.getByTitle("New message").click();

  await pageA.getByPlaceholder("Search people...").fill(userB.email);
  await pageA.getByRole("button", { name: new RegExp(userB.email) }).click();
  await pageA.getByRole("button", { name: "Start chat" }).click();

  await expect(pageA.getByPlaceholder("Message...")).toBeVisible();

  const messageContent = `Hello from Alice ${crypto.randomUUID()}`;
  await pageA.getByPlaceholder("Message...").fill(messageContent);

  const createMessageResponsePromise = pageA.waitForResponse((response) => {
    const url = new URL(response.url());

    return (
      response.request().method() === "POST" &&
      url.pathname.endsWith("/chats/messages")
    );
  });

  await pageA.getByPlaceholder("Message...").press("Enter");
  const createMessageResponse = await createMessageResponsePromise;
  expect(createMessageResponse.ok()).toBe(true);

  await pageB.getByRole("button", { name: "Messages" }).click();
  await pageB.getByRole("button", { name: new RegExp(userA.name) }).click();

  await expect(pageB.getByText(messageContent)).toBeVisible({
    timeout: 15_000,
  });

  await pageA.context().close();
  await pageB.context().close();
});

test("project collaborators can start a group message and all recipients receive it", async ({
  browser,
  request,
}) => {
  const userA = await registerUser(request, backendURL, {
    name: "Alice Group E2E",
  });
  const userB = await registerUser(request, backendURL, {
    name: "Bob Group E2E",
  });
  const userC = await registerUser(request, backendURL, {
    name: "Carol Group E2E",
  });

  const pageA = await loginAsUser(browser, userA);
  const pageB = await loginAsUser(browser, userB);
  const pageC = await loginAsUser(browser, userC);
  const projectName = `E2E Group Message Project ${crypto.randomUUID()}`;

  await pageA
    .getByRole("banner")
    .getByRole("button", { name: "Create project" })
    .click();
  const createProjectDialog = pageA.getByRole("dialog", {
    name: "Create project",
  });
  await createProjectDialog.locator("#name").fill(projectName);
  await createProjectDialog
    .locator("#description")
    .pressSequentially("Created by the group-message e2e test");
  await createProjectDialog
    .getByRole("button", { name: "Create project" })
    .click();
  await expectToast(pageA, "Project created successfully");

  await pageA.getByRole("link", { name: projectName }).click();
  await pageA.getByTitle("Add project member").click();
  let addMemberDialog = pageA.getByRole("dialog", {
    name: "Add project member",
  });
  await addMemberDialog.getByLabel("Email").fill(userB.email);
  await addMemberDialog.getByRole("button", { name: "Add member" }).click();
  await expectToast(pageA, "Member added successfully");

  await pageA.getByTitle("Add project member").click();
  addMemberDialog = pageA.getByRole("dialog", { name: "Add project member" });
  await addMemberDialog.getByLabel("Email").fill(userC.email);
  await addMemberDialog.getByRole("button", { name: "Add member" }).click();
  await expectToast(pageA, "Member added successfully");

  await pageA.goto("/projects");
  await pageA.getByRole("button", { name: "Messages" }).click();
  await pageA.getByTitle("New message").click();
  const peopleSearch = pageA.getByPlaceholder("Search people...");
  await peopleSearch.fill(userB.email);
  await pageA.getByRole("button", { name: new RegExp(userB.email) }).click();
  await peopleSearch.fill(userC.email);
  await pageA.getByRole("button", { name: new RegExp(userC.email) }).click();
  await pageA.getByRole("button", { name: "Start chat" }).click();

  const messageContent = `Hello group ${crypto.randomUUID()}`;
  await pageA.getByPlaceholder("Message...").fill(messageContent);
  await pageA.getByPlaceholder("Message...").press("Enter");

  for (const recipientPage of [pageB, pageC]) {
    await recipientPage.getByRole("button", { name: "Messages" }).click();
    await recipientPage
      .getByText("2 members", { exact: true })
      .locator("xpath=ancestor::button[1]")
      .click();
    await expect(recipientPage.getByText(messageContent)).toBeVisible({
      timeout: 15_000,
    });
  }

  await pageA.context().close();
  await pageB.context().close();
  await pageC.context().close();
});

import crypto from "node:crypto";
import { test, expect } from "../src/fixtures/authenticated-page.js";

test.describe("auth", () => {
  test("unauthenticated user is redirected away from protected pages", async ({
    page,
  }) => {
    await page.goto("/projects");

    await expect(page).toHaveURL(/\/login$/);
    await expect(
      page.getByRole("heading", { name: "Welcome back" })
    ).toBeVisible();
  });

  test("registered user can log in via the UI", async ({ page, testUser }) => {
    await page.goto("/login");

    await page.locator("#email").fill(testUser.email);
    await page.locator("#password").fill(testUser.password);
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(
      page.getByText(`Welcome back, ${testUser.name}`)
    ).toBeVisible();
    await expect(page).toHaveURL(/\/projects$/);
  });

  test("authenticated session survives a reload and logout revokes protected access", async ({
    authenticatedPage: page,
    testUser,
  }) => {
    await page.reload();

    await expect(page).toHaveURL(/\/projects$/);
    await expect(
      page.getByText(`Welcome back, ${testUser.name}`)
    ).toBeVisible();

    await page.getByRole("banner").locator("button").last().click();
    await page.getByRole("menuitem", { name: "Logout" }).click();

    await expect(page).toHaveURL(/\/login$/);
    await expect(
      page.getByRole("heading", { name: "Welcome back" })
    ).toBeVisible();

    await page.goto("/projects");

    await expect(page).toHaveURL(/\/login$/);
  });

  test("user can change their password and use only the new password to log in", async ({
    authenticatedPage: page,
    testUser,
  }) => {
    const newPassword = "UpdatedPassword123!";

    await page.getByRole("banner").locator("button").last().click();
    await page.getByRole("menuitem", { name: "Change password" }).click();

    const changePasswordDialog = page.getByRole("dialog", {
      name: "Change password",
    });
    await changePasswordDialog
      .getByLabel("Current password")
      .fill(testUser.password);
    await changePasswordDialog
      .getByLabel("New password", { exact: true })
      .fill(newPassword);
    await changePasswordDialog
      .getByLabel("Confirm new password")
      .fill(newPassword);
    await changePasswordDialog
      .getByRole("button", { name: "Update password" })
      .click();

    await expect(page.getByText("Password updated")).toBeVisible();
    await expect(changePasswordDialog).toHaveCount(0);

    await page.getByRole("banner").locator("button").last().click();
    await page.getByRole("menuitem", { name: "Logout" }).click();

    await page.locator("#email").fill(testUser.email);
    await page.locator("#password").fill(testUser.password);
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(page).toHaveURL(/\/login$/);

    await page.locator("#password").fill(newPassword);
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(page).toHaveURL(/\/projects$/);
    await expect(
      page.getByText(`Welcome back, ${testUser.name}`)
    ).toBeVisible();
  });

  test("wrong password is rejected and stays on the login page", async ({
    page,
    testUser,
  }) => {
    await page.goto("/login");

    await page.locator("#email").fill(testUser.email);
    await page.locator("#password").fill("definitely-the-wrong-password");
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(page).toHaveURL(/\/login$/);
  });

  test("newly signed-up user can log in via the UI", async ({ page }) => {
    const email = `e2e-signup-${crypto.randomUUID()}@example.com`;
    const name = "New E2E User";
    const password = "Password123!";

    await page.goto("/sign-up");

    await page.locator("#name").fill(name);
    await page.locator("#email").fill(email);
    await page.locator("#password").fill(password);
    await page.locator("#confirmPassword").fill(password);
    await page.getByRole("button", { name: "Sign up" }).click();

    await expect(page).toHaveURL(/\/login$/);

    await page.locator("#email").fill(email);
    await page.locator("#password").fill(password);
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(page.getByText(`Welcome back, ${name}`)).toBeVisible();
    await expect(page).toHaveURL(/\/projects$/);
  });
});

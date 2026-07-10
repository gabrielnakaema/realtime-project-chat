import crypto from "node:crypto";
import { test, expect } from "../src/fixtures/authenticated-page.js";

test.describe("auth", () => {
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

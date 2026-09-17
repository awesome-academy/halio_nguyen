import { test, expect } from "@playwright/test";

/**
 * Zero network calls in this file — the assertion that no request fires is
 * itself the point. Messages are read verbatim from
 * `apps/web/src/lib/validation/auth-schema.ts` (both fields are
 * `z.string().min(1, ...)`; there is no `.email()` format check, so a
 * syntactically odd email is NOT rejected client-side — see the second
 * test's comment for why this file does not include a "malformed email"
 * case).
 */
test.describe("login client-side validation", () => {
  test("empty submit shows both required-field errors and sends no request", async ({ page }) => {
    let loginCalls = 0;
    await page.route("**/api/v1/admin/auth/login", (route) => {
      loginCalls += 1;
      return route.abort();
    });

    await page.goto("/admin/login");
    await page.getByRole("button", { name: "Sign In" }).click();

    // The form is noValidate, so these come from zod, not the browser.
    await expect(page.getByText("Email is required")).toBeVisible();
    await expect(page.getByText("Password is required")).toBeVisible();
    await expect(page.getByRole("button", { name: "Sign In" })).toBeEnabled();
    await expect(page).toHaveURL(/\/admin\/login$/);
    expect(loginCalls, "empty submit must not reach the API").toBe(0);
  });

  test("email filled, password empty shows only the password error and sends no request", async ({ page }) => {
    let loginCalls = 0;
    await page.route("**/api/v1/admin/auth/login", (route) => {
      loginCalls += 1;
      return route.abort();
    });

    await page.goto("/admin/login");
    // auth-schema.ts requires only a non-empty string for email — no format
    // check exists in production code, and this plan is read-only there, so
    // a "not-an-email" scenario would actually pass validation and fire a
    // real (unbudgeted) login request. This value only needs to be non-empty.
    await page.getByLabel("Email").fill("someone@example.com");
    await page.getByRole("button", { name: "Sign In" }).click();

    await expect(page.getByText("Password is required")).toBeVisible();
    await expect(page.getByText("Email is required")).not.toBeVisible();
    await expect(page).toHaveURL(/\/admin\/login$/);
    expect(loginCalls, "partial submit must not reach the API").toBe(0);
  });
});

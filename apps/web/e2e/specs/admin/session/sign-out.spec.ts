import { test, expect } from "../../../fixtures/test";

/**
 * Clearing the cookie here affects only this test's own browser context; the
 * on-disk `storageState` file is untouched, so parallel workers replaying it
 * are unaffected (decisions.md §J4 risk table).
 */
test("Sign Out ends the session and re-arms the middleware gate", async ({ page, shell }) => {
  await page.goto("/admin/dashboard");
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();

  await shell.signOut.click();
  // Match on pathname only, not the whole href: onSettled's plain
  // router.replace("/admin/login") can race the global 401 handler — the
  // dashboard's still-mounted session query re-fires against the
  // just-cleared cookie before the cache clear finishes unmounting it, and
  // that handler's own replace appends `?from=/admin/dashboard`. Either way
  // sign-out has already succeeded server-side and the destination is the
  // same login page, so the query string is incidental, not asserted.
  await page.waitForURL((url) => url.pathname === "/admin/login");
  await expect(page.getByRole("heading", { level: 1, name: "SUN Admin" })).toBeVisible();

  // The server cleared the cookie unconditionally, so the middleware gate is
  // armed again for this context. middleware.ts percent-encodes the `from`
  // value, so the query param is compared parsed, not as a raw string/regex
  // against the encoded href.
  await page.goto("/admin/tours");
  await page.waitForURL((url) => url.pathname === "/admin/login" && url.searchParams.get("from") === "/admin/tours");
});

import { test, expect } from "../../../fixtures/test";

/**
 * Failure injection via `page.route`, not production code (decisions.md
 * §J9) — the global 401 handler in `AdminShell` is a high-value path that is
 * unreachable against a healthy stack, so both cases here fulfil a real
 * request with a fabricated 401 rather than touching any production file.
 */
test.describe("global 401 handling", () => {
  test("a 401 from a list endpoint sends the admin back to login with ?from", async ({ page }) => {
    await page.goto("/admin/dashboard");
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();

    await page.route("**/api/v1/admin/tours**", (route) =>
      route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: { code: "unauthorized", message: "unauthorized" } }),
      }),
    );

    await page.getByRole("link", { name: "Tour Packages", exact: true }).click();
    await expect(page).toHaveURL(/\/admin\/login\?from=/);
  });

  // Note: a 401 injected on `**/api/v1/admin/auth/me` while ALREADY on
  // `/admin/login` is a no-op in the real app — `useAdminSession` in
  // `AdminShell` is called with `enabled: !isLoginRoute`, so that request is
  // never issued on the login route at all (verified in
  // `admin-shell.tsx`/`use-admin-session.ts`). Asserting against a request
  // that never fires would be a fake pass, so this case instead exercises
  // the handler's pathname guard the way the real app can actually trigger
  // it: a 401 on the session-hydration call from an authenticated page,
  // followed by a check that landing on /admin/login is stable (no further
  // bounce), which is exactly what the guard in `admin-shell.tsx` exists to
  // prevent.
  test("a 401 from the session-hydration call redirects once and does not loop", async ({ page }) => {
    await page.route("**/api/v1/admin/auth/me", (route) =>
      route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: { code: "unauthorized", message: "unauthorized" } }),
      }),
    );

    await page.goto("/admin/users");
    // AdminShell builds the target with encodeURIComponent, so the query
    // param is compared parsed, not as a raw string/regex against the
    // encoded href (decisions.md Risk Assessment: "?from= regex brittle
    // against URL encoding").
    const landedOnLoginForUsers = (url: URL) =>
      url.pathname === "/admin/login" && url.searchParams.get("from") === "/admin/users";
    await page.waitForURL(landedOnLoginForUsers);

    // Stable, not bouncing: still on /admin/login, and the login form itself
    // is fully interactive — proof the pathname guard stopped a second
    // redirect rather than merely racing one that hasn't landed yet.
    await expect(page.getByRole("heading", { level: 1, name: "SUN Admin" })).toBeVisible();
    expect(landedOnLoginForUsers(new URL(page.url()))).toBe(true);
    await expect(page.getByRole("button", { name: "Sign In" })).toBeEnabled();
  });
});

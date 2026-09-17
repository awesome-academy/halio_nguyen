import { test, expect } from "@playwright/test";
import { env } from "../../config/env";

/**
 * The middleware redirect gate (`apps/web/src/middleware.ts`). It checks
 * cookie PRESENCE only, never validity — the last test in this file proves
 * that directly. Costs zero real login calls: every case here runs with an
 * empty or garbage-cookie context.
 */
const GUARDED = [
  "/admin/dashboard",
  "/admin/tours",
  "/admin/categories",
  "/admin/bookings",
  "/admin/reviews",
  "/admin/users",
  "/admin/revenue",
];

test.describe("middleware access gate", () => {
  for (const path of GUARDED) {
    test(`redirects ${path} to login carrying ?from`, async ({ page }) => {
      await page.goto(path);
      // middleware.ts builds the target with URLSearchParams.set, which
      // percent-encodes the "/" in the from value (e.g. `?from=%2Fadmin%2F...`).
      // Compare the parsed query param, not a raw string/regex against the
      // encoded href (decisions.md Risk Assessment: "?from= regex brittle
      // against URL encoding").
      await page.waitForURL((url) => url.pathname === "/admin/login" && url.searchParams.get("from") === path);
      await expect(page.getByRole("heading", { level: 1, name: "SUN Admin" })).toBeVisible();
    });
  }

  test("login page itself is reachable without a cookie", async ({ page }) => {
    await page.goto("/admin/login");
    await expect(page).toHaveURL(/\/admin\/login$/);
    await expect(page.getByText("Sign in to the management console")).toBeVisible();
  });

  test("middleware checks cookie PRESENCE only, not validity", async ({ page, context }) => {
    await context.addCookies([
      { name: "sun_admin_token", value: "not-a-real-jwt", url: env.baseURL },
    ]);
    const response = await page.goto("/admin/tours");
    // Assert on the actual navigation response, not a UI poll: once
    // /admin/tours has already been compiled by an earlier test, the
    // client-side 401 redirect can fire faster than a `toHaveURL` poll would
    // ever observe the pre-redirect state, making that a race rather than a
    // real assertion. `page.goto`'s response reflects the middleware's HTTP
    // decision before any client JS runs — a server-side redirect changes
    // its final URL, so a same-page 200 here proves the middleware let the
    // request through despite the invalid cookie.
    expect(response?.status(), "middleware must not reject this at the HTTP layer").toBe(200);
    expect(new URL(response!.url()).pathname, "middleware must not redirect this at the HTTP layer").toBe(
      "/admin/tours",
    );
    // ...but the API rejects the garbage token and the AdminShell global 401
    // handler takes over, client-side, afterwards.
    await expect(page).toHaveURL(/\/admin\/login/, { timeout: 20_000 });
  });
});

import { readFileSync } from "node:fs";
import { test as setup, expect } from "@playwright/test";
import { env, AUTH_STATE_PATH } from "../config/env";

setup("authenticate as admin and capture storageState", async ({ page, context }) => {
  // Login-budget evidence (mandatory override, phase-03): count every real
  // POST /auth/login this run makes, so the total is measured, not claimed.
  page.on("response", (res) => {
    if (res.request().method() === "POST" && res.url().endsWith("/api/v1/admin/auth/login")) {
      console.log(`[e2e-login-call] ${res.status()} ${res.url()}`);
    }
  });

  // Phase-03 amendment: carry a `?from=` the way middleware.ts would set it on
  // a real unauthenticated hit, so the one successful admin login in the
  // whole suite also proves the redirect round trip (LoginForm reads
  // searchParams.get("from") and router.replace(redirect_to) honours it)
  // instead of costing a second real login elsewhere.
  await page.goto("/admin/login?from=/admin/categories");
  await expect(page.getByRole("heading", { level: 1, name: "SUN Admin" })).toBeVisible();
  await expect(page.getByText("Sign in to the management console")).toBeVisible();

  await page.getByLabel("Email").fill(env.adminEmail);
  await page.getByLabel("Password").fill(env.adminPassword);
  await page.getByRole("button", { name: "Sign In" }).click();

  // The form does router.replace(redirect_to) with the validated `from` value.
  // Matched on the URL's pathname, not a string-suffix regex: the `from`
  // value itself is "/admin/categories", so a suffix match against the full
  // href would false-positive while still sitting on
  // /admin/login?from=/admin/categories (decisions.md Risk Assessment,
  // "?from= regex brittle against URL encoding").
  await page.waitForURL((url) => url.pathname === "/admin/categories");
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();

  // Move on to the dashboard before capturing state: every authenticated spec
  // expects storageState to represent a plain authenticated session, not one
  // mid-redirect-round-trip.
  await page.goto("/admin/dashboard");
  await expect(page).toHaveURL(/\/admin\/dashboard$/);
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();

  // Liveness probe (decisions §J2): proves dev server + /api/v1 rewrite + Go API are
  // all alive. A stale server on :3000 most often fails exactly here.
  const me = await page.request.get("/api/v1/admin/auth/me");
  expect(me.status(), "GET /auth/me after login").toBe(200);

  await context.storageState({ path: AUTH_STATE_PATH });

  // AC-4: prove empirically that the HttpOnly cookie was captured. The whole
  // authenticated suite rests on this; it is asserted, never assumed.
  const state = JSON.parse(readFileSync(AUTH_STATE_PATH, "utf8")) as {
    cookies: { name: string; httpOnly: boolean; path: string; expires: number }[];
  };
  const token = state.cookies.find((c) => c.name === "sun_admin_token");
  expect(token, "sun_admin_token missing from storageState").toBeDefined();
  expect(token!.httpOnly, "sun_admin_token should be HttpOnly").toBe(true);
  expect(token!.path).toBe("/");

  // Cookie Max-Age is 1 hour; a run longer than that will start failing mid-suite.
  // Record the capture time alongside the storageState so later runs/specs can
  // fail fast with an explicit "session expired" message instead of surfacing
  // as scattered 401s (mandatory override #3).
  const capturedAtPath = `${AUTH_STATE_PATH}.captured-at`;
  const { writeFileSync } = await import("node:fs");
  writeFileSync(capturedAtPath, String(Date.now()), "utf8");

  const minutesLeft = Math.round((token!.expires * 1000 - Date.now()) / 60_000);
  console.log(`[e2e] session captured, ~${minutesLeft} min of validity remaining`);
  expect(minutesLeft, "session too short to run the suite").toBeGreaterThan(10);
});

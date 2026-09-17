import { test, expect } from "@playwright/test";
import { env } from "../../config/env";
import { formAlert } from "../../support/form-alert";

// Serial: keeps these three real POST /auth/login calls apart in time. The API
// enforces a 5/min per-IP bucket AND a 5/min-per-email throttle (decisions.md
// §J6, corrected — see the §J5 addendum for why this file is a fourth serial
// block). All three attempts land well inside either limit.
test.describe.configure({ mode: "serial" });

test.describe("login failures", () => {
  // Captured from the first real failure and reused below — never guessed
  // (decisions.md Risk Assessment: "Alert text guessed and hardcoded").
  let capturedAlertText = "";

  // Login-budget evidence (mandatory override): count every real
  // POST /auth/login this file makes, so the total is measured, not claimed.
  test.beforeEach(async ({ page }) => {
    page.on("response", (res) => {
      if (res.request().method() === "POST" && res.url().endsWith("/api/v1/admin/auth/login")) {
        console.log(`[e2e-login-call] ${res.status()} ${res.url()}`);
      }
    });
  });

  test("wrong password shows ONE form-level alert and clears the password", async ({ page, context }) => {
    await page.goto("/admin/login");
    // tourist@, never admin@ — the admin email's throttle window belongs to setup.
    await page.getByLabel("Email").fill(env.userEmail);
    await page.getByLabel("Password").fill("definitely-not-the-password");
    await page.getByRole("button", { name: "Sign In" }).click();

    const alert = formAlert(page);
    await expect(alert).toBeVisible();
    await expect(alert).toHaveCount(1);
    // Dev-mode StrictMode can render this alert before its text settles —
    // `not.toBeEmpty()` is itself a retrying web-first assertion, so it waits
    // out that flicker instead of racing a one-shot `textContent()` read.
    await expect(alert).not.toBeEmpty();
    capturedAlertText = (await alert.textContent()) ?? "";
    expect(capturedAlertText.length, "alert must carry real text").toBeGreaterThan(0);

    // BR-001: never field-level — that would leak which field failed.
    await expect(page.getByLabel("Email")).toHaveValue(env.userEmail);
    await expect(page.getByLabel("Password")).toHaveValue("");
    await expect(page).toHaveURL(/\/admin\/login/);

    // A rejected login never sets the session cookie.
    const cookies = await context.cookies();
    expect(cookies.find((c) => c.name === "sun_admin_token")).toBeUndefined();
  });

  test("unknown account gives the same message (no user enumeration)", async ({ page }) => {
    await page.goto("/admin/login");
    await page.getByLabel("Email").fill("e2e-nobody@example.invalid");
    await page.getByLabel("Password").fill("Whatever@123456");
    await page.getByRole("button", { name: "Sign In" }).click();

    const alert = formAlert(page);
    await expect(alert).toBeVisible();
    await expect(alert).toHaveText(capturedAlertText);
    await expect(page).toHaveURL(/\/admin\/login/);
  });

  // Mandatory override (RBAC coverage gap): a genuine non-admin account, with
  // its REAL correct password (tourist@sunbooking.com shares the seed's
  // dev-only bcrypt hash with admin@ — apps/api/migrations/000002_seed_data.up.sql
  // — so `env.adminPassword` is a correct credential here, never a guess and
  // never the literal seed password hardcoded in this file).
  //
  // Verified against source (auth_service.go): `isAdminActive := user != nil
  // && user.Role == domain.RoleAdmin && user.IsActive`. A non-admin fails that
  // check and gets the exact same generic `invalidCredentialsMessage` as a
  // wrong password or an unknown account — the SAME 401 at LOGIN, never a
  // distinct response. That means `RequireAdminRole`'s 403 branch
  // ("This account does not have admin access.", auth.go:49) is unreachable
  // through any real login: a non-admin never receives a signed session
  // token to present to that gate in the first place. This test asserts the
  // real, verified behaviour rather than the originally assumed 403 — see
  // the implementer's handback report for the full note on why the
  // gate-level branch could not also be exercised without a production
  // change or a new token-signing devDependency.
  test("a real non-admin account is rejected at LOGIN with the same generic message (role gate is unreachable)", async ({
    page,
    context,
  }) => {
    await page.goto("/admin/login");
    await page.getByLabel("Email").fill(env.userEmail);
    await page.getByLabel("Password").fill(env.adminPassword);
    await page.getByRole("button", { name: "Sign In" }).click();

    const alert = formAlert(page);
    await expect(alert).toBeVisible();
    await expect(alert).toHaveText(capturedAlertText);
    await expect(page).toHaveURL(/\/admin\/login/);

    // The decisive proof: correct credentials, wrong role, still no cookie.
    const cookies = await context.cookies();
    expect(cookies.find((c) => c.name === "sun_admin_token")).toBeUndefined();
  });
});

import { test, expect } from "@playwright/test";
import { env } from "../../config/env";
import { formAlert } from "../../support/form-alert";

// Running this burns the per-IP bucket for ~3 minutes and will fail every later
// login, including auth.setup.ts. Opt in, and run it ALONE:
//   $env:E2E_RATE_LIMIT_PROBE="1"; pnpm test:e2e --project=anon -g "rate limit"
test.skip(!env.rateLimitProbe, "rate-limit probe is opt-in — see decisions.md §J6");

test("six rapid bad logins trip the 429 limiter", async ({ page }) => {
  await page.goto("/admin/login");
  for (let i = 0; i < 6; i += 1) {
    await page.getByLabel("Email").fill(`e2e-burst-${i}@example.invalid`);
    await page.getByLabel("Password").fill("Whatever@123456");
    await page.getByRole("button", { name: "Sign In" }).click();
    await expect(formAlert(page)).toBeVisible();
  }
  await expect(formAlert(page)).toContainText(/too many login attempts/i);
});

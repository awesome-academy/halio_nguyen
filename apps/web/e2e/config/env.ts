import { config } from "dotenv";

// Playwright runs with cwd = apps/web. dotenv's default `import "dotenv/config"`
// only looks for `.env`, so the path is explicit here to actually pick up
// `.env.e2e` (decisions.md §J10, phase-01 step 6).
config({ path: ".env.e2e" });

function required(key: string): string {
  const value = process.env[key];
  if (!value) {
    throw new Error(
      `[e2e] Missing ${key}. Copy apps/web/.env.e2e.example to apps/web/.env.e2e and fill it in.`,
    );
  }
  return value;
}

export const env = {
  baseURL: process.env.E2E_BASE_URL ?? "http://localhost:3000",
  adminEmail: required("E2E_ADMIN_EMAIL"),
  adminPassword: required("E2E_ADMIN_PASSWORD"),
  /** Non-admin seed account, used only for deliberate bad-login tests (decisions §J6). */
  userEmail: process.env.E2E_USER_EMAIL ?? "tourist@sunbooking.com",
  /** Opt-in: the 429 probe poisons the rate-limit bucket, so it must run alone. */
  rateLimitProbe: process.env.E2E_RATE_LIMIT_PROBE === "1",
};

export const AUTH_STATE_PATH = "playwright/.auth/admin.json";

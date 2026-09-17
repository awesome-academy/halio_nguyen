import { defineConfig, devices } from "@playwright/test";
import { env, AUTH_STATE_PATH } from "./e2e/config/env";

const isCI = !!process.env.CI;

export default defineConfig({
  testDir: "./e2e/specs",
  // Parallelism is per FILE, not per test — decisions.md §J4.
  fullyParallel: false,
  workers: isCI ? 1 : 2,
  retries: isCI ? 2 : 0,
  forbidOnly: isCI,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  reporter: [
    ["list"],
    ["html", { outputFolder: "e2e-report", open: "never" }],
    // Machine-checkable evidence — decisions.md §J8. This file (and the
    // traces/videos alongside it) is credential-bearing: it can contain
    // Set-Cookie / session data captured during the run. Keep it under
    // e2e-report/ (gitignored) and never upload it anywhere.
    ["json", { outputFile: "e2e-report/results.json" }],
  ],
  use: {
    baseURL: env.baseURL,
    // Dev-mode first-hit compilation is slow; this is deliberate — decisions.md §J1.
    navigationTimeout: 90_000,
    actionTimeout: 15_000,
    // NOTE: traces/videos/screenshots land in test-results/ (gitignored) and
    // can capture Set-Cookie headers and the live session JWT. Never upload
    // these as CI artifacts or share them outside the local machine.
    trace: "retain-on-failure",
    video: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "setup",
      testDir: "./e2e/setup",
      testMatch: /.*\.setup\.ts/,
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "anon",
      testDir: "./e2e/specs/anon",
      use: { ...devices["Desktop Chrome"], storageState: undefined },
    },
    {
      name: "admin",
      testDir: "./e2e/specs/admin",
      dependencies: ["setup"],
      use: { ...devices["Desktop Chrome"], storageState: AUTH_STATE_PATH },
    },
    // Widening the matrix is a config change, never a rewrite — decisions.md §J3.
    // { name: "admin-firefox", testDir: "./e2e/specs/admin", dependencies: ["setup"],
    //   use: { ...devices["Desktop Firefox"], storageState: AUTH_STATE_PATH } },
  ],
  webServer: {
    // MUST go through the package script: it runs clean-next-dir.mjs, which does not
    // follow the standalone symlinks. Never `next dev`/`next build` directly, never
    // `rm -rf .next` — build-machine hazard #2.
    command: process.env.PW_WEB_MODE === "prod" ? "pnpm build && pnpm start" : "pnpm dev",
    url: `${env.baseURL}/admin/login`,
    reuseExistingServer: !isCI && process.env.PW_FRESH_SERVER !== "1",
    timeout: 180_000,
    stdout: "pipe",
    stderr: "pipe",
  },
});

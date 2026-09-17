import { existsSync, readFileSync } from "node:fs";
import { AUTH_STATE_PATH } from "./env";

/** Cookie Max-Age is 1 hour (decisions.md, phase-01 Key Insights). */
const SESSION_MAX_AGE_MS = 60 * 60 * 1000;
/** Same safety margin auth.setup.ts itself enforces right after capture. */
const MIN_REMAINING_MS = 10 * 60 * 1000;

export const CAPTURED_AT_PATH = `${AUTH_STATE_PATH}.captured-at`;

/**
 * Fail-fast guard against the mandatory-override session-expiry hazard: the
 * `sun_admin_token` cookie is only valid for 1 hour, and `middleware.ts`
 * checks its *presence*, not its validity — so a stale `storageState` gets
 * past the middleware and only fails later as a scattered, hard-to-diagnose
 * 401 deep inside an unrelated spec. Call this once per authenticated spec
 * file (e.g. from a shared fixture's setup) to turn that into one clear,
 * early, actionable error instead.
 *
 * Throws when the captured-at marker is missing (setup never ran) or the
 * session is expired / expiring within `MIN_REMAINING_MS`.
 */
export function assertSessionFresh(now: number = Date.now()): void {
  if (!existsSync(CAPTURED_AT_PATH)) {
    throw new Error(
      `[e2e] No captured session found at ${CAPTURED_AT_PATH}. ` +
        `Run: pnpm test:e2e --project=setup`,
    );
  }

  const capturedAt = Number(readFileSync(CAPTURED_AT_PATH, "utf8"));
  if (!Number.isFinite(capturedAt)) {
    throw new Error(
      `[e2e] Captured session marker at ${CAPTURED_AT_PATH} is corrupt. ` +
        `Re-run: pnpm test:e2e --project=setup`,
    );
  }

  const remainingMs = SESSION_MAX_AGE_MS - (now - capturedAt);
  if (remainingMs < MIN_REMAINING_MS) {
    const elapsedMin = Math.round((now - capturedAt) / 60_000);
    throw new Error(
      `[e2e] session expired — re-run \`pnpm test:e2e --project=setup\` ` +
        `(captured ${elapsedMin} min ago, cookie Max-Age is 60 min).`,
    );
  }
}

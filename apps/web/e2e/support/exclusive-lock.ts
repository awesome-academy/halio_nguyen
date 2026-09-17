import { existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const LOCK_ROOT = join(tmpdir(), "sun-booking-e2e-locks");
/** A lock older than this is assumed abandoned by a crashed/killed run. */
const STALE_MS = 5 * 60 * 1000;
const POLL_MS = 500;
const WAIT_TIMEOUT_MS = 60_000;

interface LockOwner {
  pid: number;
  at: number;
}

function lockDir(name: string): string {
  return join(LOCK_ROOT, name);
}

/** `mkdirSync` without `recursive` is an atomic create-or-fail at the OS level. */
function tryAcquire(name: string): boolean {
  try {
    mkdirSync(lockDir(name));
    writeFileSync(join(lockDir(name), "owner.json"), JSON.stringify({ pid: process.pid, at: Date.now() } satisfies LockOwner), "utf8");
    return true;
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "EEXIST") return false;
    throw error;
  }
}

function reclaimIfStale(name: string): void {
  const dir = lockDir(name);
  const ownerFile = join(dir, "owner.json");
  if (!existsSync(ownerFile)) return;
  try {
    const owner = JSON.parse(readFileSync(ownerFile, "utf8")) as LockOwner;
    if (Date.now() - owner.at > STALE_MS) rmSync(dir, { recursive: true, force: true });
  } catch {
    // Corrupt marker — treat it as abandoned and reclaim.
    rmSync(dir, { recursive: true, force: true });
  }
}

/**
 * Cross-process mutex over the shared, un-reset Postgres instance (D2).
 *
 * `test.describe.serial` (decisions.md §J5) only orders tests within one
 * Playwright process. Nothing stops a second terminal, or a second
 * developer, from running the suite concurrently against the same
 * database — two runs could both enter a serial block and one clobbers the
 * other's mid-mutation state while both report green. This lock closes that
 * gap with no new dependency: `fs.mkdirSync` is an atomic create-or-fail
 * primitive, so exactly one process wins the race.
 *
 * Fails fast with a clear message once `WAIT_TIMEOUT_MS` elapses, rather
 * than silently racing another run.
 */
export async function withExclusiveLock<T>(name: string, fn: () => Promise<T>): Promise<T> {
  mkdirSync(LOCK_ROOT, { recursive: true });
  const deadline = Date.now() + WAIT_TIMEOUT_MS;

  for (;;) {
    reclaimIfStale(name);
    if (tryAcquire(name)) break;
    if (Date.now() > deadline) {
      throw new Error(`[e2e] another e2e run holds the "${name}" lock — wait for it to finish or investigate a stale lock at ${lockDir(name)}`);
    }
    await new Promise((resolve) => setTimeout(resolve, POLL_MS));
  }

  try {
    return await fn();
  } finally {
    rmSync(lockDir(name), { recursive: true, force: true });
  }
}

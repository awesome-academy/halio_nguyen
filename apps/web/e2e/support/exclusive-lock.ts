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

/**
 * Every write to the `categories` table must take this one lock.
 *
 * The reorder write path recomputes ranks as a contiguous `0..n-1` set
 * (`category_rules.go`), so ANY concurrent category create/delete — from a
 * different spec file on the other worker — shifts the ranks the sort-order
 * test just read, and its next click lands on the wrong absolute slot.
 * Serialising all category writes against each other removes that
 * interference. See the reorder-staleness note in the plan's Execution Log.
 */
export const CATEGORY_WRITE_LOCK = "categories-write";

/**
 * Every mutation against the `users` table (role change, active toggle) must
 * take this lock — record AND restore included, not just the click.
 *
 * There are exactly two seeded users and no admin-UI path to create a third
 * (phase-06-users.md). Without this lock, a second concurrent suite run can
 * read `tourist@sunbooking.com`'s row mid-mutation from a different worker
 * process, record that transient state as its own "original", and then
 * "restore" the account to it — permanently leaving a stray admin or a
 * deactivated account behind once both runs report green.
 */
export const USER_WRITE_LOCK = "users-write";

/**
 * Locks this process already holds, so a nested acquire is a no-op instead of
 * a self-deadlock: the sort-order spec holds the lock across its whole body
 * and calls the category factory — which takes the same lock — inside it.
 * Per-process is the correct grain; each Playwright worker is its own process
 * and runs its tests sequentially.
 */
const heldInThisProcess = new Set<string>();

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
  // Reentrant: this process already owns it, so just run the body. Releasing
  // here would hand the lock away while the outer holder is still mid-write.
  if (heldInThisProcess.has(name)) return await fn();

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

  heldInThisProcess.add(name);
  try {
    return await fn();
  } finally {
    heldInThisProcess.delete(name);
    rmSync(lockDir(name), { recursive: true, force: true });
  }
}

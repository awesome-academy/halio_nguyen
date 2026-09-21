import { type Page } from "@playwright/test";
import { UsersPage } from "../pages/users-page";
import { safeTeardown } from "../support/safe-teardown";
import { USER_WRITE_LOCK, withExclusiveLock } from "../support/exclusive-lock";
import { env } from "../config/env";

type Role = "Admin" | "User";

/**
 * Record-and-revert guard for `tourist@sunbooking.com`, the only non-admin
 * seeded user (phase-06-users.md). There is no admin-UI path to create a
 * user and no reseed path that respects D2, so this account is irreplaceable
 * — every mutation against it MUST be put back, including on failure/timeout.
 *
 * The role-change control is DETAIL-PAGE ONLY (`users-columns.tsx` explicitly
 * omits it from the row; `user-detail-client.tsx`'s "Change role" button is
 * the real trigger) — restore() drives that page, never a row action.
 */
export class SeededUserGuard {
  private originalRole: Role | null = null;
  private originalActive: boolean | null = null;

  constructor(private readonly page: Page) {}

  get email(): string {
    return env.userEmail;
  }

  /** Records the OBSERVED starting state — never assumes the canonical seed. */
  async record(): Promise<void> {
    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      const users = new UsersPage(this.page);
      await users.goto(`?search=${encodeURIComponent(this.email)}`);
      this.originalRole = await users.rowRole(this.email);
      this.originalActive = await users.rowIsActive(this.email);
    });
  }

  /**
   * Restores role and active state to what `record()` observed, then VERIFIES
   * the restore actually took. The verification step is deliberately left
   * OUTSIDE `safeTeardown`: a swallowed restore failure is exactly how this
   * account would end up permanently corrupted without anyone noticing, so
   * this one step throws instead of logging.
   */
  async restore(): Promise<void> {
    if (this.originalRole === null || this.originalActive === null) return;
    const targetRole = this.originalRole;
    const targetActive = this.originalActive;

    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      await safeTeardown(`user ${this.email} role`, async () => {
        const users = new UsersPage(this.page);
        await users.goto(`?search=${encodeURIComponent(this.email)}`);
        const current = await users.rowRole(this.email);
        if (current === targetRole) return;
        await users.openDetail(this.email);
        await users.changeRoleButton.click();
        await users.setRole(targetRole);
      });

      await safeTeardown(`user ${this.email} active`, async () => {
        const users = new UsersPage(this.page);
        await users.goto(`?search=${encodeURIComponent(this.email)}`);
        const isActive = await users.rowIsActive(this.email);
        if (isActive === targetActive) return;
        const control = targetActive ? users.activate(this.email) : users.deactivate(this.email);
        await control.click();
        await users.confirmToggle(targetActive ? "Activate" : "Deactivate");
      });

      // Noisy on purpose (see class doc): a mismatch here fails the test run
      // rather than leaving `tourist@` silently corrupted for the next one.
      const users = new UsersPage(this.page);
      await users.goto(`?search=${encodeURIComponent(this.email)}`);
      const finalRole = await users.rowRole(this.email);
      const finalActive = await users.rowIsActive(this.email);
      if (finalRole !== targetRole || finalActive !== targetActive) {
        throw new Error(
          `[e2e] FAILED TO RESTORE ${this.email}: expected role=${targetRole} active=${targetActive}, ` +
            `found role=${finalRole} active=${finalActive}. Manual repair required — see decisions.md/phase-06 runbook.`,
        );
      }
    });
  }
}

import { test, expect } from "../../../fixtures/test";
import { UsersPage } from "../../../pages/users-page";
import { env } from "../../../config/env";
import { USER_WRITE_LOCK, withExclusiveLock } from "../../../support/exclusive-lock";

/**
 * Own-account protection (BR-001 usability affordance). No mutation — this
 * is real business logic that happens to be free to test.
 */
test.describe("own-account protection", () => {
  test("the signed-in admin's own row offers no actions", async ({ page }) => {
    const users = new UsersPage(page);
    await users.goto(`?search=${encodeURIComponent(env.adminEmail)}`);
    await users.table.expectRowVisible(env.adminEmail);

    await expect(users.deactivate(env.adminEmail)).toHaveCount(0);
    await expect(users.activate(env.adminEmail)).toHaveCount(0);
    await expect(users.deleteControl(env.adminEmail)).toHaveCount(0);
  });

  // Holds USER_WRITE_LOCK despite mutating nothing: the row renders
  // "Deactivate {email}" only while the account is active, and
  // `users-reversible.spec.ts` deactivates it for part of its run on the
  // other worker.
  test("the other user's row DOES offer actions (control for the test above)", async ({ page }) => {
    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      const users = new UsersPage(page);
      await users.goto(`?search=${encodeURIComponent(env.userEmail)}`);
      await expect(users.deactivate(env.userEmail)).toBeVisible();
      // Present and reachable — NEVER clicked anywhere in this suite. See phase-06-users.md.
      await expect(users.deleteControl(env.userEmail)).toBeVisible();
    });
  });

  test("own detail page states the account cannot be changed here", async ({ page }) => {
    const users = new UsersPage(page);
    await users.goto(`?search=${encodeURIComponent(env.adminEmail)}`);
    await users.openDetail(env.adminEmail);
    await expect(users.ownAccountNotice).toBeVisible();
    await expect(users.deleteUserButton).toHaveCount(0);
    await expect(users.changeRoleButton).toHaveCount(0);
  });
});

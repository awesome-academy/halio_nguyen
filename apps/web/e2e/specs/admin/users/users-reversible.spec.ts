import { test, expect } from "../../../fixtures/test";
import { UsersPage } from "../../../pages/users-page";
import {
  USER_WRITE_LOCK,
  withExclusiveLock,
} from "../../../support/exclusive-lock";

// SERIAL: every test here mutates the ONE non-admin seeded user
// (tourist@sunbooking.com). Concurrent runs would see each other's
// half-applied state (decisions.md §J5). `USER_WRITE_LOCK` additionally
// guards against a second, fully separate `playwright test` invocation on
// this machine (a cross-PROCESS race `describe.serial` cannot reach).
// The 180 s budget is wall-clock, not leniency on any assertion: each test
// here mutates through the detail page and the `seededUser` teardown then
// re-navigates to restore AND verify the restore. Under `pnpm dev` (§J1) with
// two workers sharing one server, on-demand route compilation pushes that
// past the 60 s default — and a teardown killed by a timeout is precisely how
// this irreplaceable account would be left corrupted.
test.describe.configure({ mode: "serial", timeout: 180_000 });

test.describe("reversible user mutations", () => {
  test("role can be changed to Admin and back, within one lock window", async ({
    page,
    seededUser,
  }) => {
    const users = new UsersPage(page);

    // Assert AND revert inside the lock. Releasing it before the revert left a
    // window where this user was Admin while the lock was free, so a
    // lock-holding read on the other worker (users-list's Role filter) saw a
    // legitimately-Admin row and failed. The fixture teardown stays as the
    // safety net for the failure path; on the happy path it finds no drift.
    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      await users.goto(`?search=${encodeURIComponent(seededUser.email)}`);
      await users.openDetail(seededUser.email);
      await users.changeRoleButton.click();
      await expect(users.roleDialog).toBeVisible();
      await users.setRole("Admin");

      await users.goto(`?search=${encodeURIComponent(seededUser.email)}`);
      await expect(users.table.rowByText(seededUser.email)).toContainText(
        /admin/i,
      );

      await users.openDetail(seededUser.email);
      await users.changeRoleButton.click();
      await expect(users.roleDialog).toBeVisible();
      await users.setRole("User");

      await users.goto(`?search=${encodeURIComponent(seededUser.email)}`);
      await expect(users.table.rowByText(seededUser.email)).not.toContainText(
        /admin/i,
      );
    });
  });

  test("the Change role dialog can be cancelled with no effect", async ({
    page,
    seededUser,
  }) => {
    const users = new UsersPage(page);

    // Cancel issues no mutation of its own, but this test COMPARES the row
    // before and after — so a concurrent mutation (another suite run on this
    // machine) would show up as a false "Cancel had an effect". The lock is
    // for the comparison, not the click.
    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      await users.goto(`?search=${encodeURIComponent(seededUser.email)}`);
      const before = await users.table.rowByText(seededUser.email).innerText();

      await users.openDetail(seededUser.email);
      await users.changeRoleButton.click();
      await expect(users.roleDialog).toBeVisible();
      await users.roleDialog.getByRole("button", { name: "Cancel" }).click();
      await expect(users.roleDialog).toHaveCount(0);

      await users.goto(`?search=${encodeURIComponent(seededUser.email)}`);
      // Compare like with like: innerText() keeps the cell tab separators that
      // toHaveText() normalises away, so the two representations never match.
      // expect.poll keeps the retry behaviour a web-first assertion would give.
      await expect
        .poll(() => users.table.rowByText(seededUser.email).innerText(), {
          timeout: 10_000,
        })
        .toBe(before);
    });
  });

  test("deactivate then reactivate the other user", async ({
    page,
    seededUser,
  }) => {
    const users = new UsersPage(page);

    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      await users.goto(`?search=${encodeURIComponent(seededUser.email)}`);

      await users.deactivate(seededUser.email).click();
      await users.confirmToggle("Deactivate");
      await expect(users.activate(seededUser.email)).toBeVisible();

      await users.activate(seededUser.email).click();
      await users.confirmToggle("Activate");
      await expect(users.deactivate(seededUser.email)).toBeVisible();
    });
    // Back to active — seededUser's teardown will find no drift to restore.
  });
});

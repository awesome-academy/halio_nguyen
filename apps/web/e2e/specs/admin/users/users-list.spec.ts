import { test, expect } from "../../../fixtures/test";
import { UsersPage } from "../../../pages/users-page";
import { pickOption } from "../../../components/radix-select";
import { env } from "../../../config/env";
import { USER_WRITE_LOCK, withExclusiveLock } from "../../../support/exclusive-lock";

/** Read-only: list, search, filters, empty state, detail route. No mutation — safe to run in parallel. */
test.describe("users list", () => {
  test("lists both seeded users", async ({ page }) => {
    const users = new UsersPage(page);
    await users.goto();
    await users.table.expectRowVisible(env.adminEmail);
    await users.table.expectRowVisible(env.userEmail);
  });

  test("search by email narrows to one user", async ({ page }) => {
    const users = new UsersPage(page);
    await users.goto();
    await users.table.search(env.userEmail);
    await expect(page).toHaveURL(/[?&]search=/);
    await users.table.expectRowVisible(env.userEmail);
    await users.table.expectRowGone(env.adminEmail);
  });

  test("a no-match search shows the empty message", async ({ page }) => {
    const users = new UsersPage(page);
    await users.goto("?search=E2E_no_such_person_zzz");
    await expect(users.table.emptyMessage("No users match your search.")).toBeVisible();
  });

  // Reads role state, so it must hold USER_WRITE_LOCK even though it mutates
  // nothing: `users-reversible.spec.ts` temporarily promotes the other user to
  // Admin, and on the other worker that makes this row legitimately present
  // under role=admin — failing expectRowGone for a reason that is not a bug.
  // A lock is the honest fix; asserting less would hide the filter regression
  // this test exists to catch.
  test("Role filter narrows the list and lands in the URL", async ({ page }) => {
    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      const users = new UsersPage(page);
      await users.goto();
      await pickOption(page, "Role filter", "Admin");
      await expect(page).toHaveURL(/[?&]role=admin/);
      await users.table.expectRowVisible(env.adminEmail);
      await users.table.expectRowGone(env.userEmail);
    });
  });

  // Same reason as the Role filter above: `users-reversible.spec.ts`
  // temporarily deactivates this user, which would legitimately hide the row
  // from an Active filter running on the other worker.
  test("Status filter lands in the URL", async ({ page }) => {
    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      const users = new UsersPage(page);
      await users.goto();
      await pickOption(page, "Status filter", "Active");
      await expect(page).toHaveURL(/[?&]is_active=true/);
      await users.table.expectRowVisible(env.userEmail);
    });
  });

  test("the detail route renders for the other user", async ({ page }) => {
    const users = new UsersPage(page);
    await users.goto(`?search=${encodeURIComponent(env.userEmail)}`);
    await users.openDetail(env.userEmail);
    // The email renders twice on the detail page — once as the PageHeader
    // description <p>, once as a <dd> in the profile list — so a bare
    // getByText is a strict-mode violation. Assert the data field itself.
    await expect(page.getByRole("definition").filter({ hasText: env.userEmail })).toBeVisible();
  });

  // Runtime backstop for the delete ban (mandatory override #4): a grep is
  // whitespace-sensitive and defeatable by reformatting, so this asserts the
  // live counts directly. No filter/search narrows this — it is the raw total.
  // Counts admins, so it holds USER_WRITE_LOCK for the same reason the two
  // filter tests above do: while `users-reversible.spec.ts` has the other user
  // promoted, "of 1" is legitimately "of 2" and this invariant would report a
  // delete-ban breach that never happened.
  test("exactly two users exist and exactly one is an admin (runtime invariant)", async ({ page }) => {
    await withExclusiveLock(USER_WRITE_LOCK, async () => {
      const users = new UsersPage(page);
      await users.goto();
      await expect(users.table.rangeSummary).toHaveText(/of 2$/);

      await pickOption(page, "Role filter", "Admin");
      await expect(users.table.rangeSummary).toHaveText(/of 1$/);
    });
  });
});

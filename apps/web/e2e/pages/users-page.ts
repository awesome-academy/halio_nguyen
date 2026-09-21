import { expect, type Page } from "@playwright/test";
import { AdminPage } from "./admin-page";

/**
 * `/admin/users` list plus the `/admin/users/{id}` detail route.
 *
 * One POM covers both routes (rather than a second class) because every
 * detail-page control here is addressed by role/aria-label alone and the
 * list/detail screens share the same `DataTableComponent`/`ConfirmDialogComponent`
 * base (`AdminPage`). `goto()` only knows the list path; detail navigation
 * goes through `gotoDetail` or by following the row's link, matching how a
 * real admin reaches it.
 */
export class UsersPage extends AdminPage {
  constructor(page: Page) {
    super(page, "/admin/users", "User Management");
  }

  // --- List row actions -----------------------------------------------
  // `users-columns.tsx`: hidden entirely on the signed-in admin's own row.

  deactivate(email: string) {
    return this.page.getByRole("button", { name: `Deactivate ${email}` });
  }

  activate(email: string) {
    return this.page.getByRole("button", { name: `Activate ${email}` });
  }

  /**
   * Present for assertion ONLY. This control is never clicked anywhere in
   * this suite: only 2 users exist, the admin UI cannot create one, and D2
   * forbids reseeding a deleted row. See phase-06-users.md.
   */
  deleteControl(email: string) {
    return this.page.getByRole("button", { name: `Delete ${email}` });
  }

  /** Confirms the `Deactivate {email}`/`Activate {email}` row action's ConfirmDialog. */
  async confirmToggle(nextLabel: "Deactivate" | "Activate"): Promise<void> {
    await expect(this.confirm.root).toBeVisible();
    await this.confirm.confirm(nextLabel);
  }

  async gotoDetail(id: string): Promise<void> {
    await this.page.goto(`/admin/users/${id}`);
  }

  /** Follows the row's email link to `/admin/users/{id}` (users-columns.tsx). */
  async openDetail(email: string): Promise<void> {
    const link = this.table.rowByText(email).getByRole("link").first();
    await expect(link).toBeVisible();
    await link.click();
    // Click ONCE, then wait. `/admin/users/[id]` is compiled on demand by
    // `pnpm dev` (decisions.md §J1), and with `workers: 2` sharing one dev
    // server that first hit can exceed the default 10 s expect timeout.
    // Re-clicking to "help" aborts the in-flight navigation and makes this
    // strictly worse — a patient wait is the fix, not a retry.
    await expect(this.page).toHaveURL(/\/admin\/users\/[0-9a-f-]+$/, { timeout: 30_000 });
  }

  // --- Detail page: role change (FR-005, detail-only — SCR002) ---------

  get changeRoleButton() {
    return this.page.getByRole("button", { name: "Change role" });
  }

  get roleDialog() {
    return this.page.getByRole("dialog", { name: "Change role" });
  }

  get roleSelect() {
    return this.roleDialog.getByRole("combobox", { name: "Role" });
  }

  async setRole(role: "Admin" | "User"): Promise<void> {
    await this.roleSelect.click();
    await this.page.getByRole("option", { name: role, exact: true }).click();
    await this.roleDialog.getByRole("button", { name: "Save" }).click();
    await expect(this.roleDialog).toHaveCount(0);
  }

  // --- Detail page: active toggle + own-account guard ------------------

  get toggleActive() {
    return this.page.getByRole("switch", { name: "Toggle account active" });
  }

  get ownAccountNotice() {
    return this.page.getByText("You can't change your own account from here.");
  }

  get deleteUserButton() {
    return this.page.getByRole("button", { name: "Delete user" });
  }

  /** Reads the role badge text ("Admin" | "User") out of a list row. */
  async rowRole(email: string): Promise<"Admin" | "User"> {
    const text = await this.table.rowByText(email).innerText();
    return /admin/i.test(text) ? "Admin" : "User";
  }

  /** `Deactivate {email}` present ⇒ the row is currently active. */
  async rowIsActive(email: string): Promise<boolean> {
    return (await this.deactivate(email).count()) > 0;
  }
}

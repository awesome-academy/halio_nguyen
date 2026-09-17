import { expect, type Page } from "@playwright/test";
import { DataTableComponent } from "../components/data-table";
import { ConfirmDialogComponent } from "../components/confirm-dialog";

/**
 * POM base shared by every admin list screen: the `<h1>` heading, the
 * DataTable, its delete confirm dialog, and the shared error panel/Retry
 * button pattern used when a list query fails (see decisions.md §J9).
 *
 * Deliberately has no toast wiring — `ToastComponent` was cut (YAGNI): it
 * had exactly two call sites, both in one later-phase spec file, so that
 * spec inlines `page.getByRole("status").filter({ hasText })` directly.
 */
export abstract class AdminPage {
  readonly table: DataTableComponent;
  readonly confirm: ConfirmDialogComponent;

  protected constructor(
    readonly page: Page,
    readonly path: string,
    readonly heading: string,
  ) {
    this.table = new DataTableComponent(page);
    this.confirm = new ConfirmDialogComponent(page);
  }

  async goto(query = ""): Promise<void> {
    await this.page.goto(`${this.path}${query}`);
    await this.expectLoaded();
  }

  async expectLoaded(): Promise<void> {
    await expect(this.page.getByRole("heading", { level: 1, name: this.heading })).toBeVisible();
  }

  /** Shared list error panel: "Could not load {resource}." plus a Retry button. */
  errorPanel(resource: string) {
    return this.page.getByText(`Could not load ${resource}.`, { exact: true });
  }

  get retryButton() {
    return this.page.getByRole("button", { name: "Retry" });
  }
}

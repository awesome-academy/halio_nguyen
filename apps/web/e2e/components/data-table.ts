import { expect, type Locator, type Page } from "@playwright/test";

/**
 * Wraps the shared `DataTable`/`DataTableToolbar`/`DataTablePagination` used
 * by every admin list screen (categories, tours, bookings, reviews, users).
 */
export class DataTableComponent {
  constructor(private readonly page: Page) {}

  get table(): Locator {
    return this.page.getByRole("table");
  }

  /**
   * The ONLY sanctioned way to address a row. Never index rows positionally
   * and never assert a raw row count — the DB is shared with other workers
   * and other developers (decisions.md §J4).
   */
  rowByText(text: string): Locator {
    return this.table.getByRole("row").filter({ hasText: text });
  }

  cellsOf(text: string): Locator {
    return this.rowByText(text).getByRole("cell");
  }

  get searchBox(): Locator {
    return this.page.getByRole("textbox", { name: "Search" });
  }

  /** 300ms debounce (data-table-toolbar.tsx): fill, then assert on settled UI — no sleeping. */
  async search(term: string): Promise<void> {
    await this.searchBox.fill(term);
  }

  sortButton(columnTitle: string): Locator {
    return this.page.getByRole("button", { name: `Sort by ${columnTitle}` });
  }

  async sortBy(columnTitle: string): Promise<void> {
    await this.sortButton(columnTitle).click();
  }

  columnHeader(title: string): Locator {
    return this.page.getByRole("columnheader", { name: title });
  }

  /** "Rows per page" is both visible text AND the trigger's aria-label — role is required to disambiguate. */
  get rowsPerPage(): Locator {
    return this.page.getByRole("combobox", { name: "Rows per page" });
  }

  async setRowsPerPage(size: 10 | 20 | 50 | 100): Promise<void> {
    await this.rowsPerPage.click();
    await this.page.getByRole("option", { name: String(size), exact: true }).click();
  }

  get nextPage(): Locator {
    return this.page.getByRole("button", { name: "Next page" });
  }

  get previousPage(): Locator {
    return this.page.getByRole("button", { name: "Previous page" });
  }

  /** Summary uses an EN DASH (U+2013): "1–20 of 42". Match the stable tail only. */
  get rangeSummary(): Locator {
    return this.page.getByText(/\d+.\d+ of \d+/);
  }

  pageSummary(page: number, pageCount: number): Locator {
    return this.page.getByText(`Page ${page} of ${pageCount}`, { exact: true });
  }

  emptyMessage(message: string): Locator {
    return this.page.getByText(message, { exact: true });
  }

  /** Row-scoped action trigger, e.g. `Actions for {title}` (tours) or an icon button's aria-label. */
  rowActions(ariaLabel: string): Locator {
    return this.page.getByRole("button", { name: ariaLabel });
  }

  async openRowActions(ariaLabel: string): Promise<void> {
    await this.rowActions(ariaLabel).click();
  }

  /** Radix `DropdownMenuItem` renders into a portal, so this is page-scoped, not row-scoped. */
  menuItem(name: string): Locator {
    return this.page.getByRole("menuitem", { name });
  }

  async expectRowVisible(text: string): Promise<void> {
    await expect(this.rowByText(text)).toBeVisible();
  }

  async expectRowGone(text: string): Promise<void> {
    await expect(this.rowByText(text)).toHaveCount(0);
  }
}

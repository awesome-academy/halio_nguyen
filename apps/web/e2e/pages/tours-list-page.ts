import { type Locator, type Page } from "@playwright/test";
import { AdminPage } from "./admin-page";

export type TourRowAction = "Edit" | "Publish" | "Archive" | "Reactivate" | "Delete";

/** `/admin/tours` — list plus the row-menu status/delete actions (`tour-row-actions.tsx`). */
export class ToursListPage extends AdminPage {
  constructor(page: Page) {
    super(page, "/admin/tours", "Tour Packages");
  }

  // A <Link> inside <Button asChild> — role is LINK, not button. Unlike categories'
  // "New Category", which is a real <button> (decisions.md's Corrections §2).
  get newTourLink(): Locator {
    return this.page.getByRole("link", { name: "New Tour" });
  }

  rowMenu(title: string): Locator {
    return this.page.getByRole("button", { name: `Actions for ${title}` });
  }

  async openRowMenu(title: string): Promise<void> {
    await this.rowMenu(title).click();
  }

  /** Radix `DropdownMenuItem` renders into a portal, so this is page-scoped, not row-scoped. */
  action(name: TourRowAction): Locator {
    return this.page.getByRole("menuitem", { name });
  }

  async runAction(title: string, name: TourRowAction): Promise<void> {
    await this.openRowMenu(title);
    await this.action(name).click();
  }
}

import { expect, type Page } from "@playwright/test";
import { TourFormPage, type TourBasicInput } from "../pages/tour-form-page";
import { AdminPage } from "../pages/admin-page";
import { uniqueName } from "../support/unique-name";
import { safeTeardown } from "../support/safe-teardown";

/** A seeded, always-present category — used so `tourFactory` never depends on `categoryFactory`. */
const DEFAULT_CATEGORY = "Island & Coastal Tours";

class ToursListPage extends AdminPage {
  constructor(page: Page) {
    super(page, "/admin/tours", "Tour Packages");
  }
}

export type TourInput = Partial<TourBasicInput> & { title?: string };

/**
 * Self-cleaning tour factory (D2). Creates the minimum valid payload
 * `tourCreateSchema` accepts — images and schedules are optional arrays, so
 * the form's Basic + Pricing sections alone are enough.
 *
 * Deletion goes through the tours row action menu: `Actions for {title}` →
 * `Delete` → the shared alertdialog → `Delete`. New tours are created in
 * `draft` status (`tour_rules.go`), and `delete` is only offered in that
 * status — a factory-created tour that gets published/archived by the test
 * body will fail its own cleanup, which `safeTeardown` turns into a logged
 * orphan rather than a masked test failure.
 */
export class TourFactory {
  private readonly created: string[] = [];

  constructor(private readonly page: Page) {}

  async create(overrides: TourInput = {}): Promise<string> {
    const title = overrides.title ?? uniqueName("Tour");
    const form = new TourFormPage(this.page);
    await form.gotoNew();
    await form.fillBasic({
      title,
      categoryName: overrides.categoryName ?? DEFAULT_CATEGORY,
      destination: overrides.destination ?? "Da Nang, Vietnam",
      description: overrides.description ?? "E2E-created tour for kit verification.",
      durationDays: overrides.durationDays ?? 3,
      durationNights: overrides.durationNights ?? 2,
      price: overrides.price ?? 1_000_000,
      maxParticipants: overrides.maxParticipants ?? 20,
    });
    await form.submitCreate.click();
    await expect(this.page).toHaveURL(/\/admin\/tours\/[0-9a-f-]{36}$/);
    this.created.push(title);
    return title;
  }

  async cleanup(): Promise<void> {
    for (const title of [...this.created].reverse()) {
      await safeTeardown(`tour ${title}`, async () => {
        const tours = new ToursListPage(this.page);
        await tours.goto(`?search=${encodeURIComponent(title)}`);
        await tours.table.rowActions(`Actions for ${title}`).click();
        await tours.table.menuItem("Delete").click();
        await tours.confirm.confirm("Delete");
        await tours.table.expectRowGone(title);
      });
    }
  }
}

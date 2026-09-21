import { expect, type Page } from "@playwright/test";
import { TourFormPage, type TourBasicInput } from "../pages/tour-form-page";
import { uniqueName } from "../support/unique-name";
import { safeTeardown } from "../support/safe-teardown";

/** A seeded, always-present category — used so `tourFactory` never depends on `categoryFactory`. */
const DEFAULT_CATEGORY = "Island & Coastal Tours";

export type TourInput = Partial<TourBasicInput> & { title?: string };

/**
 * Self-cleaning tour factory (D2). Creates the minimum valid payload
 * `tourCreateSchema` accepts — images and schedules are optional arrays, so
 * the form's Basic + Pricing sections alone are enough.
 *
 * Cleanup deletes via the admin API rather than the row action menu, because
 * the UI can paint itself into a corner this factory must still clean up:
 * `allowedTourActions` (tour-status-rules.ts) offers `delete` ONLY for
 * `draft`, and the status machine has no transition back to it
 * (draft→published→archived→reactivate→published). So any tour a test
 * publishes is permanently undeletable through the UI, and the old
 * menu-driven cleanup leaked one orphan per lifecycle run — silently, since
 * `safeTeardown` swallows teardown errors so they cannot mask a real failure.
 * `DELETE /tours/:id` itself has no status check (tour_service.go:152 blocks
 * only on active bookings), so the API path cleans up every state.
 *
 * This is teardown, not an assertion: deleting a tour THROUGH the UI still
 * has its own explicit test in `tours-crud.spec.ts`. The request goes through
 * the same origin, proxy and session cookie the UI uses.
 */
export class TourFactory {
  private readonly created: { id: string; title: string }[] = [];

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
    // The create redirect carries the new id — cheaper and more reliable than
    // searching the list for it at teardown time.
    const id = new URL(this.page.url()).pathname.split("/").pop()!;
    this.created.push({ id, title });
    return title;
  }

  async cleanup(): Promise<void> {
    for (const { id, title } of [...this.created].reverse()) {
      await safeTeardown(`tour ${title}`, async () => {
        const response = await this.page.request.delete(`/api/v1/admin/tours/${id}`);
        // 404 means something already removed it — equally clean. Anything
        // else is a real leak and must be loud, because `safeTeardown` will
        // otherwise reduce it to a log line nobody reads.
        if (!response.ok() && response.status() !== 404) {
          throw new Error(`[e2e] could not delete tour ${title} (${id}): HTTP ${response.status()} ${await response.text()}`);
        }
      });
    }
  }
}

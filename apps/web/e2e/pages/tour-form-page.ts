import { expect, type Locator, type Page } from "@playwright/test";
import { pickOption } from "../components/radix-select";
import { TOUR_FIELDS, TOUR_IMAGE_FIELDS, TOUR_SCHEDULE_FIELDS } from "./tour-form-fields";

/** The minimum payload `tourCreateSchema` accepts — images/schedules are optional arrays. */
export interface TourBasicInput {
  title: string;
  categoryName: string;
  destination: string;
  description: string;
  durationDays: number;
  durationNights: number;
  price: number;
  maxParticipants: number;
}

/** `/admin/tours/new` and `/admin/tours/:id` — the create/edit form plus its two array editors. */
export class TourFormPage {
  constructor(readonly page: Page) {}

  async gotoNew(): Promise<void> {
    await this.page.goto("/admin/tours/new");
    await expect(this.page.getByRole("heading", { level: 1, name: "New tour" })).toBeVisible();
  }

  async gotoEdit(id: string): Promise<void> {
    await this.page.goto(`/admin/tours/${id}`);
  }

  get submitCreate(): Locator {
    return this.page.getByRole("button", { name: "Create tour" });
  }

  get submitSave(): Locator {
    return this.page.getByRole("button", { name: "Save changes" });
  }

  /**
   * `getByLabel` matches by substring by default, and "Price (VND)" is a
   * substring of "Discounted price (VND)" — `exact` is required so the two
   * pricing fields never collide.
   */
  field(label: string): Locator {
    return this.page.getByLabel(label, { exact: true });
  }

  /** Fills every required base field. Category is a Radix Select, not a plain input. */
  async fillBasic(input: TourBasicInput): Promise<void> {
    await this.field(TOUR_FIELDS.title).fill(input.title);
    await pickOption(this.page, TOUR_FIELDS.category, input.categoryName);
    await this.field(TOUR_FIELDS.destination).fill(input.destination);
    await this.field(TOUR_FIELDS.description).fill(input.description);
    await this.field(TOUR_FIELDS.durationDays).fill(String(input.durationDays));
    await this.field(TOUR_FIELDS.durationNights).fill(String(input.durationNights));
    await this.field(TOUR_FIELDS.price).fill(String(input.price));
    await this.field(TOUR_FIELDS.maxParticipants).fill(String(input.maxParticipants));
  }

  // --- images editor (tour-images-editor.tsx) ---------------------------
  // Save/Move are edit-mode only; create-mode rows submit with the form.
  get addImage(): Locator {
    return this.page.getByRole("button", { name: TOUR_IMAGE_FIELDS.addImage });
  }

  saveImage(n: number): Locator {
    return this.page.getByRole("button", { name: TOUR_IMAGE_FIELDS.saveImage(n) });
  }

  removeImage(n: number): Locator {
    return this.page.getByRole("button", { name: TOUR_IMAGE_FIELDS.removeImage(n) });
  }

  moveImageUp(n: number): Locator {
    return this.page.getByRole("button", { name: TOUR_IMAGE_FIELDS.moveImageUp(n) });
  }

  moveImageDown(n: number): Locator {
    return this.page.getByRole("button", { name: TOUR_IMAGE_FIELDS.moveImageDown(n) });
  }

  // --- schedules editor (tour-schedules-editor.tsx) ---------------------
  get addSchedule(): Locator {
    return this.page.getByRole("button", { name: TOUR_SCHEDULE_FIELDS.addSchedule });
  }

  saveSchedule(n: number): Locator {
    return this.page.getByRole("button", { name: TOUR_SCHEDULE_FIELDS.saveSchedule(n) });
  }

  removeSchedule(n: number): Locator {
    return this.page.getByRole("button", { name: TOUR_SCHEDULE_FIELDS.removeSchedule(n) });
  }

  get departureDate(): Locator {
    return this.page.getByLabel(TOUR_SCHEDULE_FIELDS.departureDate);
  }

  get returnDate(): Locator {
    return this.page.getByLabel(TOUR_SCHEDULE_FIELDS.returnDate);
  }

  get availableSlots(): Locator {
    return this.page.getByLabel(TOUR_SCHEDULE_FIELDS.availableSlots);
  }

  get priceOverride(): Locator {
    return this.page.getByLabel(TOUR_SCHEDULE_FIELDS.priceOverride);
  }
}

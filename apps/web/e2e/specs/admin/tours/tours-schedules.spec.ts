import { test, expect } from "../../../fixtures/test";
import { ToursListPage } from "../../../pages/tours-list-page";
import { TourFormPage } from "../../../pages/tour-form-page";

/**
 * Schedule dates are native `<input type="date">` (`tour-schedule-row.tsx`),
 * not a `react-day-picker` popover — confirmed at implementation time
 * (decisions.md's risk was a false alarm), so plain `.fill("YYYY-MM-DD")`
 * works without any calendar navigation.
 */
test("a schedule can be added with all four fields and saved, and it persists", async ({ page, tourFactory }) => {
  const title = await tourFactory.create();
  const tours = new ToursListPage(page);
  const form = new TourFormPage(page);

  await tours.goto(`?search=${encodeURIComponent(title)}`);
  await tours.runAction(title, "Edit");

  const departure = new Date(Date.now() + 30 * 86_400_000).toISOString().slice(0, 10);
  const returnDate = new Date(Date.now() + 33 * 86_400_000).toISOString().slice(0, 10);

  await form.addSchedule.click();
  await form.departureDate.fill(departure);
  await form.returnDate.fill(returnDate);
  await form.availableSlots.fill("12");
  await form.priceOverride.fill("1990000");
  // handleSave reads the row via form.getValues(...), which is unaffected
  // by the id-shadowing defect documented below — a new schedule has no
  // domain id yet, so this is a plain POST and genuinely persists.
  await form.saveSchedule(1).click();
  await expect(form.removeSchedule(1)).toBeVisible();

  await page.reload();
  await expect(form.departureDate).toHaveValue(departure);
  await expect(form.returnDate).toHaveValue(returnDate);
  await expect(form.availableSlots).toHaveValue("12");
});

/**
 * DISCOVERED DEFECT (production code is read-only for this plan, so this
 * is reported, not fixed): `handleRemove` in `tour-schedules-editor.tsx`
 * reads `.id` off `fields[index]` — react-hook-form's `useFieldArray`
 * output — as the schedule's server id. That property is actually RHF's
 * own per-row tracking key (the same shadowing bug as the images editor's
 * `handleMove`/`handleRemove`), so the DELETE call 404s. Unlike the images
 * service, `TourScheduleService.Delete` is NOT idempotent-on-no-match — it
 * locks the row first and genuinely errors when it can't find one — so the
 * failure surfaces as a toast and the row is never optimistically hidden
 * (a more honest failure mode than the images editor's, but still: a saved
 * schedule can never be removed through this screen).
 */
test("removing a saved schedule fails and the row stays (discovered defect)", async ({ page, tourFactory }) => {
  const title = await tourFactory.create();
  const tours = new ToursListPage(page);
  const form = new TourFormPage(page);

  await tours.goto(`?search=${encodeURIComponent(title)}`);
  await tours.runAction(title, "Edit");

  const departure = new Date(Date.now() + 40 * 86_400_000).toISOString().slice(0, 10);
  const returnDate = new Date(Date.now() + 43 * 86_400_000).toISOString().slice(0, 10);
  await form.addSchedule.click();
  await form.departureDate.fill(departure);
  await form.returnDate.fill(returnDate);
  await form.availableSlots.fill("5");
  await form.saveSchedule(1).click();
  await expect(form.removeSchedule(1)).toBeVisible();

  const deleteResp = page.waitForResponse((r) => r.request().method() === "DELETE" && /\/schedules\//.test(r.url()));
  await form.removeSchedule(1).click();
  const resp = await deleteResp;
  expect(resp.status()).toBe(404);
  // Not idempotent like images: the failed delete is caught and the row
  // is left in place rather than being optimistically hidden.
  await expect(form.removeSchedule(1)).toBeVisible();

  await page.reload();
  await expect(form.removeSchedule(1)).toBeVisible();
});

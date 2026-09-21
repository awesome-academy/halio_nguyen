import { test, expect } from "../../../fixtures/test";
import { ToursListPage } from "../../../pages/tours-list-page";
import { TourFormPage } from "../../../pages/tour-form-page";
import { uniqueName } from "../../../support/unique-name";

/**
 * Images seeded at CREATE time (where per-row mutations don't exist yet —
 * the whole array submits with the rest of the form, A3) so this positive
 * test is untouched by the defects documented below.
 */
test("images added at creation persist in their submitted order", async ({ page }) => {
  const title = uniqueName("TourImgOK");
  const form = new TourFormPage(page);
  await form.gotoNew();
  await form.fillBasic({
    title,
    categoryName: "Island & Coastal Tours",
    destination: "Da Nang, Vietnam",
    description: "E2E images create-time order.",
    durationDays: 2,
    durationNights: 1,
    price: 500_000,
    maxParticipants: 10,
  });
  await form.addImage.click();
  await form.imageUrlInput(1).fill("https://picsum.photos/id/10/800/600");
  await form.addImage.click();
  await form.imageUrlInput(2).fill("https://picsum.photos/id/20/800/600");
  await form.submitCreate.click();
  await expect(page).toHaveURL(/\/admin\/tours\/[0-9a-f-]{36}$/);

  await expect(form.imageUrlInput(1)).toHaveValue(/id\/10\//);
  await expect(form.imageUrlInput(2)).toHaveValue(/id\/20\//);

  const tours = new ToursListPage(page);
  await tours.goto(`?search=${encodeURIComponent(title)}`);
  await tours.runAction(title, "Delete");
  await tours.confirm.confirm("Delete");
  await tours.table.expectRowGone(title);
});

/**
 * DISCOVERED DEFECTS — production code is read-only for this plan, so
 * these are reported here, not fixed:
 *
 * 1. `handleMove`/`handleRemove` in `tour-images-editor.tsx` read `.id` off
 *    `fields[index]` — react-hook-form's `useFieldArray` output — as if it
 *    were the server's image id. It is not: RHF's own per-row tracking key
 *    lives at that same property name and shadows it. Only `handleSave`
 *    reads the real domain id correctly, via `form.getValues(...)`, which
 *    is unaffected by that shadowing.
 * 2. Net effect for MOVE: every persist PUT 404s ("Tour image not found"),
 *    is caught, and the optimistic swap is rolled back — reordering a
 *    saved image silently does nothing.
 * 3. Net effect for REMOVE: the DELETE 404s the same way, but
 *    `TourImageService.Delete` (BR: A9) treats a no-match delete as
 *    idempotent success (204) by design — so the client optimistically
 *    hides the row anyway, even though nothing was removed server-side. A
 *    reload brings the "removed" row back.
 * 4. Net effect for ADD: a freshly appended row's `field.id` is already
 *    truthy (RHF assigns one immediately), so `isSaved` reads true and the
 *    URL input is disabled from the instant it appears — a gallery can
 *    only grow at CREATE time, never afterwards.
 *
 * One tour, one narrative, `test.step`s — not separate tests — because
 * steps 2-4 all need the SAME two already-saved rows the first step seeds.
 */
test("edit-mode image mutations do not persist (discovered defects)", async ({ page }) => {
  const title = uniqueName("TourImgBug");
  const form = new TourFormPage(page);
  const tours = new ToursListPage(page);

  await test.step("seed a tour with two saved images", async () => {
    await form.gotoNew();
    await form.fillBasic({
      title,
      categoryName: "Island & Coastal Tours",
      destination: "Da Nang, Vietnam",
      description: "E2E images-editor defect tour.",
      durationDays: 2,
      durationNights: 1,
      price: 500_000,
      maxParticipants: 10,
    });
    await form.addImage.click();
    await form.imageUrlInput(1).fill("https://picsum.photos/id/10/800/600");
    await form.addImage.click();
    await form.imageUrlInput(2).fill("https://picsum.photos/id/20/800/600");
    await form.submitCreate.click();
    await expect(page).toHaveURL(/\/admin\/tours\/[0-9a-f-]{36}$/);
    await expect(form.imageUrlInput(1)).toBeDisabled(); // now-saved rows lock their URL
  });

  await test.step("Move up 404s, rolls back, and never persists", async () => {
    const movePuts = Promise.all([
      page.waitForResponse((r) => r.request().method() === "PUT" && /\/images\//.test(r.url())),
      page.waitForResponse((r) => r.request().method() === "PUT" && /\/images\//.test(r.url())),
    ]);
    await form.moveImageUp(2).click();
    const [put1, put2] = await movePuts;
    expect(put1.status()).toBe(404);
    expect(put2.status()).toBe(404);
    await expect(page.getByText("Could not reorder images.", { exact: true })).toBeVisible();
    // Rolled back: the original order is unchanged, client- and server-side.
    await expect(form.imageUrlInput(1)).toHaveValue(/id\/10\//);
    await page.reload();
    await expect(form.imageUrlInput(1)).toHaveValue(/id\/10\//);
    await expect(form.imageUrlInput(2)).toHaveValue(/id\/20\//);
  });

  await test.step("Remove looks instant client-side but does not persist", async () => {
    const removeDelete = page.waitForResponse((r) => r.request().method() === "DELETE" && /\/images\//.test(r.url()));
    await form.removeImage(2).click();
    const deleteResp = await removeDelete;
    expect(deleteResp.status()).toBe(204); // A9: no-match delete is documented-idempotent success
    await expect(form.removeImage(2)).toHaveCount(0); // optimistic client removal
    await page.reload();
    // The "removed" row is back — nothing was actually deleted server-side.
    await expect(form.removeImage(2)).toBeVisible();
    await expect(form.imageUrlInput(2)).toHaveValue(/id\/20\//);
  });

  await test.step("A freshly added row is immediately disabled", async () => {
    await form.addImage.click();
    await expect(form.imageUrlInput(3)).toBeDisabled();
  });

  await tours.goto(`?search=${encodeURIComponent(title)}`);
  await tours.runAction(title, "Delete");
  await tours.confirm.confirm("Delete");
  await tours.table.expectRowGone(title);
});

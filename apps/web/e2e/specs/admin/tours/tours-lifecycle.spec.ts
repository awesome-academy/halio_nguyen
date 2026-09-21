import { test, expect } from "../../../fixtures/test";
import { ToursListPage } from "../../../pages/tours-list-page";

/**
 * ONE test with four `test.step`s, not a `describe.serial` of separate
 * tests (mandatory override #3): `tourFactory` is test-scoped, so a serial
 * block of separate tests would tear the tour down between steps. This way
 * the single factory-created tour lives for the whole narrative and its
 * `finally` cleanup still runs once, at the end.
 *
 * `Reactivate`'s target status is read from tour-status-rules.ts /
 * tours-client.tsx (`onReactivate: (t) => changeStatus(t, "published")`),
 * not guessed: archived -> Reactivate -> published, not draft.
 */
test("tour status lifecycle: draft -> published -> archived -> reactivated", async ({ page, tourFactory }) => {
  const title = await tourFactory.create();
  const tours = new ToursListPage(page);

  await test.step("a draft offers Publish and Delete but not Archive or Reactivate", async () => {
    await tours.goto(`?search=${encodeURIComponent(title)}`);
    await expect(tours.table.rowByText(title)).toContainText(/draft/i);
    await tours.openRowMenu(title);
    await expect(tours.action("Publish")).toBeVisible();
    await expect(tours.action("Delete")).toBeVisible();
    await expect(tours.action("Archive")).toHaveCount(0);
    await expect(tours.action("Reactivate")).toHaveCount(0);
    await page.keyboard.press("Escape");
  });

  await test.step("Publish moves it to published and swaps the available actions", async () => {
    await tours.runAction(title, "Publish");
    await expect(tours.table.rowByText(title)).toContainText(/published/i);

    await tours.openRowMenu(title);
    await expect(tours.action("Archive")).toBeVisible();
    await expect(tours.action("Publish")).toHaveCount(0);
    // BR: a published tour offers no Delete — the row menu only grants it in draft.
    await expect(tours.action("Delete")).toHaveCount(0);
    await page.keyboard.press("Escape");
  });

  await test.step("Archive moves it to archived and offers only Reactivate", async () => {
    await tours.runAction(title, "Archive");
    await expect(tours.table.rowByText(title)).toContainText(/archived/i);

    await tours.openRowMenu(title);
    await expect(tours.action("Reactivate")).toBeVisible();
    await expect(tours.action("Archive")).toHaveCount(0);
    await expect(tours.action("Edit")).toHaveCount(0);
    await page.keyboard.press("Escape");
  });

  await test.step("Reactivate returns it to published and it survives a reload", async () => {
    await tours.runAction(title, "Reactivate");
    await expect(tours.table.rowByText(title)).toContainText(/published/i);
    await page.reload();
    await expect(tours.table.rowByText(title)).toContainText(/published/i);

    await tours.openRowMenu(title);
    await expect(tours.action("Archive")).toBeVisible();
    await expect(tours.action("Reactivate")).toHaveCount(0);
    await page.keyboard.press("Escape");
  });
});

import { test, expect } from "../../../fixtures/test";
import { CategoriesPage } from "../../../pages/categories-page";

// Verified against categories-columns.tsx: the ROW's own Delete button carries
// `disabled={blocked}` (blocked = tour_count > 0), not just the confirm dialog's
// `confirmDisabled`. A disabled button never fires its onClick, so the dialog
// (and its own confirmDisabled prop) is unreachable while a category has tours
// attached — the row button is the real, and only reachable, guard surface.
test.describe("category delete guard", () => {
  test("a seeded category with tours has its Delete button disabled", async ({ page }) => {
    const categories = new CategoriesPage(page);
    await categories.goto("?search=Island");

    const deleteButton = categories.deleteButton("Island & Coastal Tours");
    await expect(deleteButton).toBeDisabled();
    await expect(deleteButton).toHaveAttribute("title", /still use this category/);

    // Never opened, never touched — the row survives untouched.
    await categories.table.expectRowVisible("Island & Coastal Tours");
  });

  test("attaching a tour flips an empty category's Delete button to disabled", async ({
    page,
    categoryFactory,
    tourFactory,
  }) => {
    // This is the one path in the categories suite that also round-trips
    // through the tours form (a second, heavier route tree). Under `workers:
    // 2` with the other three category spec files running concurrently
    // against the shared dev server, observed category-list query latency
    // spikes well past the global 10s expect timeout (traced to 4s+ per
    // request under load) — genuine contention, not a race in this test's
    // own logic. `test.slow()` gives this specific test proportionally more
    // wall-clock budget rather than weakening any assertion.
    test.slow();

    const category = await categoryFactory.create();
    const categories = new CategoriesPage(page);

    // Empty category: the row's Delete button is enabled, and its dialog
    // (only now reachable) confirms deletion is permitted.
    await categories.goto(`?search=${encodeURIComponent(category)}`);
    const deleteButton = categories.deleteButton(category);
    await expect(deleteButton).toBeEnabled({ timeout: 30_000 });
    await deleteButton.click();
    await expect(categories.confirm.root).toBeVisible();
    await expect(categories.confirm.action("Delete")).toBeEnabled();
    await categories.confirm.dismiss();

    // Attach a tour: the same button must now be disabled and unclickable.
    await tourFactory.create({ categoryName: category });
    await categories.goto(`?search=${encodeURIComponent(category)}`);
    await expect(categories.deleteButton(category)).toBeDisabled({ timeout: 30_000 });
  });
});

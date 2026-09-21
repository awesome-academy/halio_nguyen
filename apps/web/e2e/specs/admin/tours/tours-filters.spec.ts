import { test, expect } from "../../../fixtures/test";
import { ToursListPage } from "../../../pages/tours-list-page";
import { pickOption } from "../../../components/radix-select";
import { uniqueName } from "../../../support/unique-name";

// Both seeded tours are `published` (DB-verified): "Phu Quoc Tropical Island
// Discovery 3D2N" in "Island & Coastal Tours", "Misty Sapa & Fansipan Peak
// Conquest 2D1N" in "Mountain & Trekking" — enough for filter/empty-state
// assertions without creating anything (Key Insight).
const PHU_QUOC = "Phu Quoc Tropical Island Discovery 3D2N";
const SAPA = "Misty Sapa & Fansipan Peak Conquest 2D1N";

test.describe("tours filters", () => {
  test("Category filter narrows the list and lands in the URL", async ({ page }) => {
    const tours = new ToursListPage(page);
    await tours.goto();
    await pickOption(page, "Category filter", "Island & Coastal Tours");
    await expect(page).toHaveURL(/[?&]category_id=/);
    await tours.table.expectRowVisible(PHU_QUOC);
    await tours.table.expectRowGone(SAPA);
  });

  test("Status filter lands in the URL and both seeded tours are published", async ({ page }) => {
    const tours = new ToursListPage(page);
    await tours.goto();
    await pickOption(page, "Status filter", "Published");
    await expect(page).toHaveURL(/[?&]status=published/);
    await tours.table.expectRowVisible(PHU_QUOC);
    await tours.table.expectRowVisible(SAPA);
  });

  test("price bounds land in the URL and exclude out-of-range tours", async ({ page }) => {
    const tours = new ToursListPage(page);
    await tours.goto();
    const minPrice = page.getByLabel("Minimum price");
    await minPrice.fill("999999999");
    // onBlur commits the bound (tours-filter-bar.tsx); Tab moves focus off
    // the field without touching production code.
    await minPrice.press("Tab");
    await expect(page).toHaveURL(/[?&]price_min=999999999/);
    await expect(tours.table.emptyMessage("No tours found.")).toBeVisible();
  });

  test("a no-match search shows the empty message", async ({ page }) => {
    const tours = new ToursListPage(page);
    await tours.goto(`?search=${encodeURIComponent(uniqueName("NoSuchTour"))}`);
    await expect(tours.table.emptyMessage("No tours found.")).toBeVisible();
  });
});

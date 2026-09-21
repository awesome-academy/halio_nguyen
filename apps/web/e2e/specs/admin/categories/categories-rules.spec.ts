import { test, expect } from "../../../fixtures/test";
import { CategoriesPage } from "../../../pages/categories-page";
import { uniqueName } from "../../../support/unique-name";
import { pickOption } from "../../../components/radix-select";

test.describe("categories business rules", () => {
  test("slug auto-follows the name until the slug is touched", async ({ page }) => {
    const categories = new CategoriesPage(page);
    await categories.goto();
    await categories.openCreate();

    const dialog = page.getByRole("dialog", { name: "New category" });
    // Verified against lib/validation/category-schema.ts's `slugify`.
    await dialog.getByLabel("Name").fill("Sunset Cruises");
    await expect(dialog.getByLabel("Slug")).toHaveValue("sunset-cruises");

    // Touch the slug: it must stop tracking from here on.
    await dialog.getByLabel("Slug").fill("my-own-slug");
    await dialog.getByLabel("Name").fill("Completely Different Name");
    await expect(dialog.getByLabel("Slug")).toHaveValue("my-own-slug");

    await categories.cancel();
  });

  test("an empty name blocks submission", async ({ page }) => {
    const categories = new CategoriesPage(page);
    await categories.goto();
    await categories.openCreate();
    await categories.save();
    // Dialog stays open — zod rejected it client-side.
    await expect(page.getByRole("dialog", { name: "New category" })).toBeVisible();
    await categories.cancel();
  });

  test("a duplicate slug surfaces as a field error, not a toast", async ({ page }) => {
    const categories = new CategoriesPage(page);
    await categories.goto();
    await categories.openCreate();
    // "island-coastal" is the REAL seeded slug for "Island & Coastal Tours"
    // (verified against the live DB — the plan's guessed
    // "island-coastal-tours" does not exist).
    await categories.fillForm({ name: uniqueName("Dup"), slug: "island-coastal" });
    await categories.save();
    // Server 409 mapped back into the form — dialog stays open, no toast.
    const dialog = page.getByRole("dialog", { name: "New category" });
    await expect(dialog).toBeVisible();
    // The footer buttons are disabled while the create mutation is in flight.
    // Under load the 409 can take longer than the default expect timeout, and
    // asserting the field error too early reads as "no error" — wait for the
    // real settle signal rather than lengthening every assertion below.
    await expect(dialog.getByRole("button", { name: "Save" })).toBeEnabled({ timeout: 30_000 });
    // Exact server message — repository.CategoryConflict: "This slug is already in use".
    await expect(dialog.getByText("This slug is already in use")).toBeVisible();
    await categories.cancel();
  });

  test("an inactive category is excluded by the Status filter", async ({ page, categoryFactory }) => {
    const name = await categoryFactory.create({ active: false });
    const categories = new CategoriesPage(page);

    await categories.goto(`?search=${encodeURIComponent(name)}`);
    await categories.table.expectRowVisible(name);

    // Status filter option is "Active", encoded on the URL as is_active=true
    // (categories-client.tsx).
    await pickOption(page, "Status filter", "Active");
    await expect(page).toHaveURL(/[?&]is_active=true/);
    await categories.table.expectRowGone(name);
  });

  test("a search with no match shows the empty message", async ({ page }) => {
    const categories = new CategoriesPage(page);
    await categories.goto(`?search=${encodeURIComponent(uniqueName("NoSuch"))}`);
    await expect(categories.table.emptyMessage("No categories found.")).toBeVisible();
  });
});

import { test, expect } from "../../../fixtures/test";
import { CategoriesPage } from "../../../pages/categories-page";
import { uniqueName } from "../../../support/unique-name";

test.describe("categories CRUD", () => {
  test("lists the seeded categories", async ({ page }) => {
    const categories = new CategoriesPage(page);
    await categories.goto();
    for (const seeded of [
      "Island & Coastal Tours",
      "Mountain & Trekking",
      "Culture & Heritage",
      "Resort & Relaxation",
    ]) {
      await categories.table.expectRowVisible(seeded);
    }
  });

  test("creates a category and it appears in the list", async ({ page, categoryFactory }) => {
    const name = await categoryFactory.create({ description: "created by the e2e suite" });
    const categories = new CategoriesPage(page);
    await categories.goto(`?search=${encodeURIComponent(name)}`);
    await categories.table.expectRowVisible(name);
  });

  test("edits a category and the row reflects the change", async ({ page, categoryFactory }) => {
    const name = await categoryFactory.create();
    const renamed = `${name}_edited`;
    const categories = new CategoriesPage(page);

    await categories.goto(`?search=${encodeURIComponent(name)}`);
    await categories.openEdit(name);
    await categories.fillForm({ name: renamed, description: "edited by the e2e suite" });
    await categories.save();
    await expect(page.getByRole("dialog")).toHaveCount(0);

    await categories.goto(`?search=${encodeURIComponent(renamed)}`);
    await categories.table.expectRowVisible(renamed);
  });

  test("deletes a category and the row disappears", async ({ page }) => {
    // Created WITHOUT the factory: this test owns the deletion, so factory teardown
    // would just log a harmless miss. Deliberate — the delete IS the assertion.
    const name = uniqueName("CatDel");
    const categories = new CategoriesPage(page);
    await categories.goto();
    await categories.openCreate();
    await categories.fillForm({ name });
    await categories.save();
    await expect(page.getByRole("dialog")).toHaveCount(0);
    await categories.table.expectRowVisible(name);

    await categories.goto(`?search=${encodeURIComponent(name)}`);
    await categories.deleteButton(name).click();
    await expect(categories.confirm.root).toBeVisible();
    await categories.confirm.confirm("Delete");
    await categories.table.expectRowGone(name);
  });
});

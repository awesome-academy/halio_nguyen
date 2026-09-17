import { expect, type Page } from "@playwright/test";
import { CategoriesPage, type CategoryInput } from "../pages/categories-page";
import { uniqueName } from "../support/unique-name";
import { safeTeardown } from "../support/safe-teardown";

/**
 * Self-cleaning category factory (D2: seed-only, no DB reset — the suite
 * creates and removes its own data through the admin UI).
 *
 * `cleanup()` is defensive by construction: it never asserts that a delete
 * succeeds. If a tour still references the category, `deleteButton` is
 * disabled by the UI (`category-delete-dialog.tsx`'s `confirmDisabled`), the
 * click/confirm sequence times out, and `safeTeardown` swallows that as a
 * logged warning instead of failing the test or the rest of the loop.
 */
export class CategoryFactory {
  private readonly created: string[] = [];

  constructor(private readonly page: Page) {}

  async create(overrides: Partial<CategoryInput> = {}): Promise<string> {
    const name = overrides.name ?? uniqueName("Cat");
    const categories = new CategoriesPage(this.page);
    await categories.goto();
    await categories.openCreate();
    await categories.fillForm({ ...overrides, name });
    await categories.save();
    await expect(this.page.getByRole("dialog")).toHaveCount(0);
    await categories.table.expectRowVisible(name);
    this.created.push(name);
    return name;
  }

  async cleanup(): Promise<void> {
    for (const name of [...this.created].reverse()) {
      await safeTeardown(`category ${name}`, async () => {
        const categories = new CategoriesPage(this.page);
        await categories.goto(`?search=${encodeURIComponent(name)}`);
        await categories.deleteButton(name).click();
        await categories.confirm.confirm("Delete");
        await categories.table.expectRowGone(name);
      });
    }
  }
}

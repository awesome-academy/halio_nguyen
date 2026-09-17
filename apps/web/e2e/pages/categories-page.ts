import { expect, type Page } from "@playwright/test";
import { AdminPage } from "./admin-page";

export interface CategoryInput {
  name: string;
  slug?: string;
  description?: string;
  imageUrl?: string;
  position?: number;
  active?: boolean;
}

/** `/admin/categories` — list plus the create/edit form dialog (`category-form-dialog.tsx`). */
export class CategoriesPage extends AdminPage {
  constructor(page: Page) {
    super(page, "/admin/categories", "Tour Categories");
  }

  // Button text is "New Category" (capital C); the dialog title is "New category"
  // (lowercase c) — decisions.md's Corrections. Scope the dialog by role, never by text.
  get newCategoryButton() {
    return this.page.getByRole("button", { name: "New Category" });
  }

  dialog(mode: "create" | "edit") {
    return this.page.getByRole("dialog", {
      name: mode === "create" ? "New category" : "Edit category",
    });
  }

  async openCreate(): Promise<void> {
    await this.newCategoryButton.click();
    await expect(this.dialog("create")).toBeVisible();
  }

  async openEdit(name: string): Promise<void> {
    await this.page.getByRole("button", { name: `Edit ${name}` }).click();
    await expect(this.dialog("edit")).toBeVisible();
  }

  async fillForm(input: CategoryInput): Promise<void> {
    const dialog = this.page.getByRole("dialog");
    await dialog.getByLabel("Name").fill(input.name);
    if (input.slug !== undefined) await dialog.getByLabel("Slug").fill(input.slug);
    if (input.description !== undefined) await dialog.getByLabel("Description").fill(input.description);
    if (input.imageUrl !== undefined) await dialog.getByLabel("Image URL").fill(input.imageUrl);
    if (input.position !== undefined) {
      await dialog.getByLabel("Position (0 = first)").fill(String(input.position));
    }
    // The switch defaults to checked (active); only toggle it when the caller wants it off.
    if (input.active === false) await dialog.getByRole("switch", { name: "Active" }).click();
  }

  async save(): Promise<void> {
    await this.page.getByRole("dialog").getByRole("button", { name: "Save" }).click();
  }

  async cancel(): Promise<void> {
    await this.page.getByRole("dialog").getByRole("button", { name: "Cancel" }).click();
  }

  moveUp(name: string) {
    return this.page.getByRole("button", { name: `Move ${name} up` });
  }

  moveDown(name: string) {
    return this.page.getByRole("button", { name: `Move ${name} down` });
  }

  /** Opens the delete confirm dialog (`category-delete-dialog.tsx`); disabled while `tour_count > 0`. */
  deleteButton(name: string) {
    return this.page.getByRole("button", { name: `Delete ${name}` });
  }
}

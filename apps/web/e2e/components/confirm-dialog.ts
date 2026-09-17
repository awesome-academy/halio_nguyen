import { expect, type Locator, type Page } from "@playwright/test";

/** Wraps the shared `ConfirmDialog` (a Radix `AlertDialog`). Cancel is always labelled "Cancel". */
export class ConfirmDialogComponent {
  constructor(private readonly page: Page) {}

  get root(): Locator {
    return this.page.getByRole("alertdialog");
  }

  title(text: string): Locator {
    return this.root.getByRole("heading", { name: text });
  }

  action(confirmLabel = "Confirm"): Locator {
    return this.root.getByRole("button", { name: confirmLabel });
  }

  get cancel(): Locator {
    return this.root.getByRole("button", { name: "Cancel" });
  }

  async confirm(confirmLabel = "Confirm"): Promise<void> {
    await this.action(confirmLabel).click();
    await expect(this.root).toHaveCount(0);
  }

  async dismiss(): Promise<void> {
    await this.cancel.click();
    await expect(this.root).toHaveCount(0);
  }
}

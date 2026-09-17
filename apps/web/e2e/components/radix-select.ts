import type { Locator, Page } from "@playwright/test";

/** Radix Select: the trigger is `role=combobox`, items are `role=option`. */
export async function pickOption(page: Page, triggerLabel: string, option: string): Promise<void> {
  await page.getByRole("combobox", { name: triggerLabel }).click();
  await page.getByRole("option", { name: option, exact: true }).click();
}

export function selectTrigger(page: Page, triggerLabel: string): Locator {
  return page.getByRole("combobox", { name: triggerLabel });
}

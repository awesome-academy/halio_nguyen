import { expect, type Locator } from "@playwright/test";

/**
 * Assert `first` appears above `second` in the table WITHOUT depending on
 * absolute positions — another worker may have inserted rows in between
 * (decisions.md §J4). Shared by the categories sort-order spec and (per
 * phase-04's plan) phase 05's tour-image ordering.
 */
export async function expectRowAbove(table: Locator, first: string, second: string): Promise<void> {
  await expect(async () => {
    const texts = await table.getByRole("row").allInnerTexts();
    const a = texts.findIndex((t) => t.includes(first));
    const b = texts.findIndex((t) => t.includes(second));
    expect(a, `${first} not found in table`).toBeGreaterThanOrEqual(0);
    expect(b, `${second} not found in table`).toBeGreaterThanOrEqual(0);
    expect(a, `${first} should sit above ${second}`).toBeLessThan(b);
  }).toPass({ timeout: 10_000 }); // explicit — toPass defaults to retrying forever
}

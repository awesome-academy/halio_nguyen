import type { Locator, Page } from "@playwright/test";

/**
 * The login form's single `role="alert"` div (BR-001: one form-level
 * message, never field-level). Plain `page.getByRole("alert")` is ambiguous
 * on every page in this app: Next.js always renders its own
 * `#__next-route-announcer__` live region with `role="alert"` too, so a
 * bare role query can resolve to either element depending on timing and
 * throws a strict-mode violation when both are present. Exclude it by id
 * rather than by content, since the announcer is legitimately empty most of
 * the time.
 */
export function formAlert(page: Page): Locator {
  return page.locator('[role="alert"]:not(#__next-route-announcer__)');
}

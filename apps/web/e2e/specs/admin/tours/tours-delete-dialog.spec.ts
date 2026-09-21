import { test, expect } from "../../../fixtures/test";
import { ToursListPage } from "../../../pages/tours-list-page";

/**
 * UI CONTRACT ONLY — this does NOT cover the server's 409 booking guard.
 * Zero bookings are seeded and the admin UI cannot create one (decisions.md
 * D5). What this proves: given a 409 shaped like the real API's, the dialog
 * renders the blocked variant (title, Close action, booking count). What it
 * does NOT prove: that the API actually returns 409 when bookings exist —
 * that server-side guard (tour_service.go's `Delete`, BR-006/BR-012) is an
 * outstanding, named, manual-verification gap (plan.md's gap list).
 *
 * The injected body mirrors the REAL envelope, verified against source, not
 * the flat shape a first draft of this test assumed:
 *   - apps/api/internal/apperror/http.go: `{ error: { code, message, fields } }`
 *   - apps/api/internal/service/tour_service.go's Delete: fields is
 *     `map[string]string`, so `booking_count` is the STRING "3", not the
 *     number 3 — Go's `map[string]string` cannot hold a numeric value.
 *   - apps/web/src/lib/api/client.ts's `toApiError`: only populates
 *     `ApiError.fields` when the body matches this nested shape.
 *   - tour-delete-dialog.tsx reads `Number(err.fields?.booking_count ?? 0)`.
 * A flat `{ message, booking_count: 3 }` body — the wrong shape — would
 * leave `ApiError.fields` `undefined` and render "0 active booking(s)",
 * silently passing for the wrong reason.
 */
test("an injected 409 flips the delete dialog to the blocked variant", async ({ page, tourFactory }) => {
  const title = await tourFactory.create();
  const tours = new ToursListPage(page);
  await tours.goto(`?search=${encodeURIComponent(title)}`);

  await page.route("**/api/v1/admin/tours/*", async (route) => {
    if (route.request().method() !== "DELETE") {
      await route.fallback();
      return;
    }
    await route.fulfill({
      status: 409,
      contentType: "application/json",
      body: JSON.stringify({
        error: {
          code: "conflict",
          message: "Cannot delete: 3 active booking(s) reference this tour.",
          fields: { booking_count: "3" },
        },
      }),
    });
  });

  await tours.runAction(title, "Delete");
  await expect(tours.confirm.root).toBeVisible();
  // Unlike a normal delete, the dialog does NOT close on this response — it
  // re-renders in place as the blocked variant (tour-delete-dialog.tsx never
  // calls onOpenChange(false) from its 409 branch), so this clicks the
  // confirm button directly rather than using the shared `.confirm()`
  // helper, which asserts the dialog closes.
  await tours.confirm.action("Delete").click();
  await expect(tours.confirm.title("Cannot delete tour")).toBeVisible();
  await expect(tours.confirm.action("Close")).toBeVisible();
  await expect(tours.confirm.root).toContainText("3");

  // Must not outlive this test: tourFactory's own teardown deletes this tour
  // through the same DELETE endpoint, and a still-active route would block
  // that too and turn a clean teardown into a false orphan.
  await tours.confirm.action("Close").click();
  await page.unroute("**/api/v1/admin/tours/*");
});

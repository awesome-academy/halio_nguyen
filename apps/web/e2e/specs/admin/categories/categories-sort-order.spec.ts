import { test, expect } from "../../../fixtures/test";
import { CategoriesPage } from "../../../pages/categories-page";
import { expectRowAbove } from "../../../support/row-order";
import { CATEGORY_WRITE_LOCK, withExclusiveLock } from "../../../support/exclusive-lock";

// SERIAL: Move up/down mutates shared sort_order across the whole table
// (CategoryService.rewriteRanks locks every non-deleted category row via
// ListForReorder's `FOR UPDATE`). Two workers reordering at once produce a
// meaningless result (decisions.md §J5).
test.describe.configure({ mode: "serial" });

// Mandatory override: `test.describe.serial` only orders tests within ONE
// Playwright process. `withExclusiveLock` additionally keeps a second,
// concurrently-running suite invocation (another terminal / developer) from
// interleaving its own reorder mutations with this file's — the cross-process
// gap `test.describe.serial` alone cannot close.
const LOCK_NAME = CATEGORY_WRITE_LOCK;

// Pin page_size=100: the default is 20 (use-list-query-state.ts), and with
// workers: 2 another spec file's freshly-created E2E_ categories can push
// these two onto page 2, where expectRowAbove (page-scoped) would never see
// them both.
const PAGE_SIZE_QS = "?page_size=100";

test.describe("category sort order", () => {
  test("Move up swaps two adjacent E2E categories and persists", async ({ page, categoryFactory }) => {
    await withExclusiveLock(LOCK_NAME, async () => {
      const first = await categoryFactory.create(); // appended at the end
      const second = await categoryFactory.create(); // appended after `first`

      const categories = new CategoriesPage(page);
      await categories.goto(PAGE_SIZE_QS);
      await expectRowAbove(categories.table.table, first, second);

      await categories.moveUp(second).click();
      await expectRowAbove(categories.table.table, second, first);

      // Persisted, not just optimistic UI.
      await page.reload();
      await categories.expectLoaded();
      await expectRowAbove(categories.table.table, second, first);
    });
  });

  test("Move down reverses it", async ({ page, categoryFactory }) => {
    await withExclusiveLock(LOCK_NAME, async () => {
      const first = await categoryFactory.create();
      const second = await categoryFactory.create();
      const categories = new CategoriesPage(page);

      await categories.goto(PAGE_SIZE_QS);
      await categories.moveDown(first).click();
      await expectRowAbove(categories.table.table, second, first);

      await page.reload();
      await expectRowAbove(categories.table.table, second, first);
    });
  });
});

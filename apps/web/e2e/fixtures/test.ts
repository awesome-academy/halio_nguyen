import { test as base, expect } from "@playwright/test";
import { assertSessionFresh } from "../config/session-guard";
import { AdminShellPage } from "../pages/admin-shell";
import { DataTableComponent } from "../components/data-table";
import { CategoryFactory } from "./category-factory";
import { TourFactory } from "./tour-factory";
import { SeededUserGuard } from "./seeded-user";

interface AdminFixtures {
  /** Fails fast, once per test, on the `admin` project only (mandatory override #6). */
  sessionGuard: void;
  shell: AdminShellPage;
  table: DataTableComponent;
  categoryFactory: CategoryFactory;
  tourFactory: TourFactory;
  /** Records `tourist@sunbooking.com`'s role/active state and restores it in `finally` (phase-06). */
  seededUser: SeededUserGuard;
}

/**
 * The single import surface for every spec (phases 03-08): `import { test,
 * expect } from "../../fixtures/test"`, never `@playwright/test` directly.
 */
export const test = base.extend<AdminFixtures>({
  // Auto-runs before every test body. Cheap (one small file read + date
  // math), and it turns a stale `sun_admin_token` (1h Max-Age) into one
  // clear, early error instead of scattered 401s in unrelated specs.
  sessionGuard: [
    async ({}, use, testInfo) => {
      if (testInfo.project.name === "admin") assertSessionFresh();
      await use();
    },
    { auto: true },
  ],

  shell: async ({ page }, use) => {
    await use(new AdminShellPage(page));
  },

  table: async ({ page }, use) => {
    await use(new DataTableComponent(page));
  },

  // try/finally, NOT afterEach — teardown must survive assertion failures
  // AND test timeouts (decisions.md §J7).
  categoryFactory: async ({ page }, use) => {
    const factory = new CategoryFactory(page);
    try {
      await use(factory);
    } finally {
      await factory.cleanup();
    }
  },

  // Depends on `categoryFactory` explicitly so Playwright's fixture graph —
  // not a spec's destructuring order — guarantees teardown runs tours
  // first, then categories (FK direction: `tours.category_id` is `ON DELETE
  // RESTRICT`). Fixture teardown is LIFO over *request* order; without this
  // explicit dependency, a spec that destructures
  // `{ tourFactory, categoryFactory }` would silently flip that order and
  // reintroduce the FK violation on cleanup.
  tourFactory: async ({ page, categoryFactory: _categoryFactory }, use) => {
    const factory = new TourFactory(page);
    try {
      await use(factory);
    } finally {
      await factory.cleanup();
    }
  },

  // try/finally, NOT afterEach — restore must survive assertion failures AND
  // test timeouts (decisions.md §J7). `record()` runs before the test body
  // touches anything so it captures the true starting state.
  seededUser: async ({ page }, use) => {
    const guard = new SeededUserGuard(page);
    await guard.record();
    try {
      await use(guard);
    } finally {
      await guard.restore();
    }
  },
});

export { expect };

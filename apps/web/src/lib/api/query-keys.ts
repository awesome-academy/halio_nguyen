import type { ListQuery } from "./types";

/**
 * Builds the one query-key shape every admin module reuses:
 * all -> lists -> list(query) and all -> details -> detail(id). Keeping
 * this in one factory (rather than six ad-hoc conventions) is what makes
 * queryClient.invalidateQueries({ queryKey: categoryKeys.lists() }) work the
 * same way for every module.
 */
function createListKeys(module: string) {
  return {
    all: [module] as const,
    lists: () => [module, "list"] as const,
    list: (query: ListQuery) => [module, "list", query] as const,
    details: () => [module, "detail"] as const,
    detail: (id: string | number) => [module, "detail", id] as const,
  };
}

// One factory per admin-portal module (F002-F007). Feature phases import
// their own factory instead of hand-rolling query keys.
export const categoryKeys = createListKeys("categories");
export const tourKeys = createListKeys("tours");
export const bookingKeys = createListKeys("bookings");
export const userKeys = createListKeys("users");
export const reviewKeys = createListKeys("reviews");
export const revenueKeys = createListKeys("revenue");

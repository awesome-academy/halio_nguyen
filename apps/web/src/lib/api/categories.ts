import { apiFetch } from "./client";
import { toQueryString } from "./query-string";
import type { ListQuery, Paginated } from "./types";
import type { Category, CategoryListItem, CategoryPayload } from "@/types/category.types";

const BASE = "/api/v1/admin/categories";

export interface CategoryListQuery extends ListQuery {
  is_active?: boolean;
}

export function listCategories(query: CategoryListQuery): Promise<Paginated<CategoryListItem>> {
  return apiFetch<Paginated<CategoryListItem>>(`${BASE}${toQueryString(query)}`);
}

export function createCategory(payload: CategoryPayload): Promise<Category> {
  return apiFetch<Category>(BASE, { method: "POST", body: payload });
}

export function updateCategory(id: string, payload: CategoryPayload): Promise<Category> {
  return apiFetch<Category>(`${BASE}/${id}`, { method: "PUT", body: payload });
}

/** ALG-001: sends the target ordinal (0-based), returns the whole reordered set. */
export function reorderCategory(id: string, sortOrder: number): Promise<{ items: Category[] }> {
  return apiFetch<{ items: Category[] }>(`${BASE}/${id}/sort-order`, {
    method: "PATCH",
    body: { sort_order: sortOrder },
  });
}

export function deleteCategory(id: string): Promise<void> {
  return apiFetch<void>(`${BASE}/${id}`, { method: "DELETE" });
}

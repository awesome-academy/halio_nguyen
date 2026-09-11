// Mirrors apps/api/internal/domain.Category and repository.CategoryListItem.

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string | null;
  image_url?: string | null;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/** One row of GET /api/v1/admin/categories — includes the derived tour_count. */
export interface CategoryListItem extends Category {
  tour_count: number;
}

/** Body of POST/PUT — the explicit DTO the API binds (never the full Category). */
export interface CategoryPayload {
  name: string;
  slug: string;
  description?: string;
  image_url?: string;
  sort_order?: number;
  is_active: boolean;
}

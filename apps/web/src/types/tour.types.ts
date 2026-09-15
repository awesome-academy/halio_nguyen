// Mirrors apps/api/internal/domain.Tour / TourImage / TourSchedule and the
// admin-only service.TourDetail / TourScheduleWithPrice shapes (Phase 4).
// Kept in one file with the pre-existing customer-facing types since no
// importer currently narrows on it (grep-verified) — admin DTOs are
// additive, nothing here removes or narrows a previously exported field.

export type TourStatus = "draft" | "published" | "archived";
export type ScheduleStatus = "open" | "closed" | "cancelled";

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  image_url?: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface TourImage {
  id: string;
  tour_id: string;
  image_url: string;
  caption?: string;
  sort_order: number;
  created_at?: string;
}

export interface TourSchedule {
  id: string;
  tour_id: string;
  departure_date: string;
  return_date: string;
  available_slots: number;
  price_override?: number;
  status: ScheduleStatus;
  created_at?: string;
  updated_at?: string;
}

/** One schedule row as returned by GET /tours/:id (service.TourScheduleWithPrice). */
export interface TourScheduleWithPrice extends TourSchedule {
  effective_price: number;
}

export interface Tour {
  id: string;
  category_id: string;
  category?: Category;
  title: string;
  slug: string;
  description: string;
  itinerary?: string;
  destination: string;
  duration_days: number;
  duration_nights: number;
  price: number;
  discount_price?: number;
  max_participants: number;
  thumbnail_url?: string;
  highlights?: string[];
  inclusions?: string;
  exclusions?: string;
  status: TourStatus;
  avg_rating: number;
  total_ratings: number;
  images?: TourImage[];
  schedules?: TourSchedule[];
  created_at: string;
  updated_at: string;
}

/** One row of GET /api/v1/admin/tours — same shape as Tour, arrays omitted. */
export type TourListItem = Tour;

/** GET /api/v1/admin/tours/:id — always carries images/schedules (service.TourDetail). */
export interface TourDetail extends Tour {
  images: TourImage[];
  schedules: TourScheduleWithPrice[];
}

/** Body shared by POST /tours and PUT /tours/:id (handler.tourRequest, minus
 * avg_rating/total_ratings/status/images/schedules which each have their own
 * dedicated path — BR-005, DEC-001, A7-A13). */
export interface TourBasePayload {
  category_id: string;
  title: string;
  description: string;
  itinerary?: string;
  destination: string;
  duration_days: number;
  duration_nights: number;
  price: number;
  discount_price?: number;
  max_participants: number;
  thumbnail_url?: string;
  highlights: string[];
  inclusions?: string;
  exclusions?: string;
}

export interface TourImagePayload {
  image_url: string;
  caption?: string;
}

/** A8's DTO — image_url is not editable once created (handler.tourImageUpdateRequest). */
export interface TourImageUpdatePayload {
  caption?: string;
  sort_order: number;
}

export interface TourSchedulePayload {
  departure_date: string;
  return_date: string;
  available_slots: number;
  price_override?: number;
}

/** POST /tours — the only place images[]/schedules[] are ever submitted (R2). */
export interface TourCreatePayload extends TourBasePayload {
  images: TourImagePayload[];
  schedules: TourSchedulePayload[];
}

/** PUT /tours/:id — base fields only; arrays and status go through A7-A13/A5. */
export type TourUpdatePayload = TourBasePayload;

import { apiFetch } from "./client";
import { toQueryString } from "./query-string";
import type { ListQuery, Paginated } from "./types";
import type {
  Tour,
  TourCreatePayload,
  TourDetail,
  TourImage,
  TourImagePayload,
  TourImageUpdatePayload,
  TourListItem,
  TourSchedule,
  TourSchedulePayload,
  TourUpdatePayload,
} from "@/types/tour.types";

const BASE = "/api/v1/admin/tours";

export interface TourListQuery extends ListQuery {
  category_id?: string;
  status?: string;
  price_min?: number;
  price_max?: number;
}

// A1
export function listTours(query: TourListQuery): Promise<Paginated<TourListItem>> {
  return apiFetch<Paginated<TourListItem>>(`${BASE}${toQueryString(query)}`);
}

// A2
export function getTour(id: string): Promise<TourDetail> {
  return apiFetch<TourDetail>(`${BASE}/${id}`);
}

// A3 — the only endpoint that accepts nested images[]/schedules[] (R2).
export function createTour(payload: TourCreatePayload): Promise<Tour> {
  return apiFetch<Tour>(BASE, { method: "POST", body: payload });
}

// A4 — base fields only; arrays and status are excluded server-side.
export function updateTour(id: string, payload: TourUpdatePayload): Promise<Tour> {
  return apiFetch<Tour>(`${BASE}/${id}`, { method: "PUT", body: payload });
}

// A5 — DEC-001's single writer of tours.status.
export function updateTourStatus(id: string, status: string): Promise<Tour> {
  return apiFetch<Tour>(`${BASE}/${id}/status`, { method: "PATCH", body: { status } });
}

// A6 — D4: 409 with { booking_count } when active bookings reference the tour.
export function deleteTour(id: string): Promise<void> {
  return apiFetch<void>(`${BASE}/${id}`, { method: "DELETE" });
}

// A7
export function createTourImage(tourId: string, payload: TourImagePayload): Promise<TourImage> {
  return apiFetch<TourImage>(`${BASE}/${tourId}/images`, { method: "POST", body: payload });
}

// A8 — caption/sort_order only (image_url is not editable once created).
export function updateTourImage(tourId: string, imageId: string, payload: TourImageUpdatePayload): Promise<TourImage> {
  return apiFetch<TourImage>(`${BASE}/${tourId}/images/${imageId}`, { method: "PUT", body: payload });
}

// A9 — idempotent: a repeat delete of an already-gone image is not an error.
export function deleteTourImage(tourId: string, imageId: string): Promise<void> {
  return apiFetch<void>(`${BASE}/${tourId}/images/${imageId}`, { method: "DELETE" });
}

// A10
export function createTourSchedule(tourId: string, payload: TourSchedulePayload): Promise<TourSchedule> {
  return apiFetch<TourSchedule>(`${BASE}/${tourId}/schedules`, { method: "POST", body: payload });
}

// A11
export function updateTourSchedule(tourId: string, scheduleId: string, payload: TourSchedulePayload): Promise<TourSchedule> {
  return apiFetch<TourSchedule>(`${BASE}/${tourId}/schedules/${scheduleId}`, { method: "PUT", body: payload });
}

// A12 — DEC-002's single writer of tour_schedules.status.
export function updateTourScheduleStatus(tourId: string, scheduleId: string, status: string): Promise<TourSchedule> {
  return apiFetch<TourSchedule>(`${BASE}/${tourId}/schedules/${scheduleId}/status`, { method: "PATCH", body: { status } });
}

// A13 — D4: 409 with { booking_count } when active bookings reference the schedule.
export function deleteTourSchedule(tourId: string, scheduleId: string): Promise<void> {
  return apiFetch<void>(`${BASE}/${tourId}/schedules/${scheduleId}`, { method: "DELETE" });
}

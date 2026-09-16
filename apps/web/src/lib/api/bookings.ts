import { apiFetch } from "./client";
import { toQueryString } from "./query-string";
import type { ListQuery, Paginated } from "./types";
import type { Booking, BookingCancelPayload, BookingDetail, BookingListItem } from "@/types/booking.types";

const BASE = "/api/v1/admin/bookings";

export interface BookingListQuery extends ListQuery {
  status?: string;
  tour_id?: string;
  schedule_id?: string;
  /** YYYY-MM-DD, inclusive, bounding bookings.created_at (FR-202). */
  date_from?: string;
  date_to?: string;
}

// A1
export function listBookings(query: BookingListQuery): Promise<Paginated<BookingListItem>> {
  return apiFetch<Paginated<BookingListItem>>(`${BASE}${toQueryString(query)}`);
}

// A2 — `payment` is absent when the booking has no payments row.
export function getBooking(id: string): Promise<BookingDetail> {
  return apiFetch<BookingDetail>(`${BASE}/${id}`);
}

// A3 — pending -> confirmed. Deliberately allowed on an unpaid booking
// (BR-002): cash and offline bank_transfer settlement depend on it, and the
// warning is the confirm dialog's job, not the server's.
export function confirmBooking(id: string): Promise<Booking> {
  return apiFetch<Booking>(`${BASE}/${id}/confirm`, { method: "PATCH" });
}

// A4 — pending|confirmed -> cancelled; restores the schedule's slots (BR-005).
export function cancelBooking(id: string, payload: BookingCancelPayload): Promise<Booking> {
  return apiFetch<Booking>(`${BASE}/${id}/cancel`, { method: "PATCH", body: payload });
}

// A5 — confirmed -> completed (BR-006).
export function completeBooking(id: string): Promise<Booking> {
  return apiFetch<Booking>(`${BASE}/${id}/complete`, { method: "PATCH" });
}

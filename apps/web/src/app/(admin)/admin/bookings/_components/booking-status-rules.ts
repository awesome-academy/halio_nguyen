import type { BookingStatus } from "@/types/booking.types";

export type BookingAction = "confirm" | "cancel" | "complete";

/**
 * SM-001, as one pure function:
 *
 *   pending ──confirm──> confirmed ──complete──> completed
 *      │                     │
 *      └──────cancel─────────┴──> cancelled   (terminal)
 *
 * This drives which buttons the UI offers. It is NOT the control — the
 * server's guarded UPDATE is (phase-06 Security Considerations); hiding a
 * button is a usability affordance only.
 */
const ALLOWED: Record<BookingStatus, BookingAction[]> = {
  pending: ["confirm", "cancel"],
  confirmed: ["complete", "cancel"],
  completed: [],
  cancelled: [],
};

export function allowedBookingActions(status: BookingStatus): BookingAction[] {
  return ALLOWED[status] ?? [];
}

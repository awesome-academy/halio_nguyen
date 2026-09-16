import { z } from "zod";

/** Mirrors maxCancellationReasonLen in
 * apps/api/internal/service/booking_service.go — keep the two in step. */
export const MAX_CANCELLATION_REASON = 1000;

// BR-004: the reason is required and length-capped. The server re-validates
// both (it trims first, so a whitespace-only reason is rejected there too).
export const bookingCancelSchema = z.object({
  cancellation_reason: z
    .string()
    .trim()
    .min(1, "A cancellation reason is required.")
    .max(MAX_CANCELLATION_REASON, `Cancellation reason must be ${MAX_CANCELLATION_REASON} characters or fewer.`),
});

export type BookingCancelFormValues = z.infer<typeof bookingCancelSchema>;

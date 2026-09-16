"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { cancelBooking, completeBooking, confirmBooking } from "@/lib/api/bookings";
import { bookingKeys, tourKeys } from "@/lib/api/query-keys";
import type { BookingCancelPayload } from "@/types/booking.types";

/**
 * One `useMutation` per A3-A5, each owning its own invalidation so no
 * component invents a cache key. Every transition changes both the row in
 * the list and the detail page, so both are invalidated.
 */

function useBookingTransition<TPayload>(mutationFn: (payload: TPayload) => Promise<unknown>, id: string, alsoInvalidateTour?: string) {
  const queryClient = useQueryClient();
  return useMutation<unknown, Error, TPayload>({
    mutationFn,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: bookingKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: bookingKeys.detail(id) });
      // Cancel restores the schedule's available_slots (BR-005), so a cached
      // tour detail showing that schedule is now stale.
      if (alsoInvalidateTour) void queryClient.invalidateQueries({ queryKey: tourKeys.detail(alsoInvalidateTour) });
    },
  });
}

export function useConfirmBooking(id: string) {
  return useBookingTransition<void>(() => confirmBooking(id), id);
}

export function useCompleteBooking(id: string) {
  return useBookingTransition<void>(() => completeBooking(id), id);
}

export function useCancelBooking(id: string, tourId?: string) {
  return useBookingTransition((payload: BookingCancelPayload) => cancelBooking(id, payload), id, tourId);
}

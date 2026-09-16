"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { useCancelBooking, useCompleteBooking, useConfirmBooking } from "@/hooks/use-booking-mutations";
import { ApiError } from "@/lib/api/types";
import type { BookingDetail } from "@/types/booking.types";
import { allowedBookingActions } from "./booking-status-rules";
import { BookingCancelDialog } from "./booking-cancel-dialog";
import { BookingConfirmDialog } from "./booking-confirm-dialog";

interface BookingStatusActionsProps {
  booking: BookingDetail;
}

/**
 * FR-401/402/403: offers exactly the transitions SM-001 permits from the
 * booking's current status. Which buttons appear is a usability affordance —
 * the server's guarded UPDATE is the actual control (FR-404), so calling a
 * transition out of order still answers 409.
 */
export function BookingStatusActions({ booking }: BookingStatusActionsProps) {
  const [openDialog, setOpenDialog] = useState<"confirm" | "cancel" | "complete" | null>(null);
  const actions = allowedBookingActions(booking.status);

  const confirmMutation = useConfirmBooking(booking.id);
  const completeMutation = useCompleteBooking(booking.id);
  const cancelMutation = useCancelBooking(booking.id, booking.tour_id);

  if (actions.length === 0) {
    return <p className="text-sm text-muted-foreground">This booking is {booking.status}; no further changes are possible.</p>;
  }

  function handleComplete() {
    completeMutation.mutate(undefined, {
      onSuccess: () => {
        toast.success("Booking marked completed");
        setOpenDialog(null);
      },
      onError: (err) => toast.error(err instanceof ApiError ? err.message : "Could not complete this booking."),
    });
  }

  return (
    <>
      <div className="flex flex-wrap items-center gap-2">
        {actions.includes("confirm") && <Button onClick={() => setOpenDialog("confirm")}>Confirm</Button>}
        {actions.includes("complete") && <Button onClick={() => setOpenDialog("complete")}>Mark completed</Button>}
        {actions.includes("cancel") && (
          <Button variant="outline" onClick={() => setOpenDialog("cancel")}>
            Cancel booking
          </Button>
        )}
      </div>

      {actions.includes("confirm") && (
        <BookingConfirmDialog
          booking={booking}
          open={openDialog === "confirm"}
          onOpenChange={(o) => setOpenDialog(o ? "confirm" : null)}
          confirmMutation={confirmMutation}
        />
      )}

      {actions.includes("complete") && (
        <ConfirmDialog
          open={openDialog === "complete"}
          onOpenChange={(o) => setOpenDialog(o ? "complete" : null)}
          title="Mark booking completed"
          confirmLabel="Mark completed"
          isPending={completeMutation.isPending}
          onConfirm={handleComplete}
          description={
            <p>
              <strong>{booking.booking_code}</strong> will be marked completed. Completed bookings cannot be changed again.
            </p>
          }
        />
      )}

      {actions.includes("cancel") && (
        <BookingCancelDialog
          booking={booking}
          open={openDialog === "cancel"}
          onOpenChange={(o) => setOpenDialog(o ? "cancel" : null)}
          cancelMutation={cancelMutation}
        />
      )}
    </>
  );
}

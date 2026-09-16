"use client";

import { AlertTriangle } from "lucide-react";
import { toast } from "sonner";
import type { UseMutationResult } from "@tanstack/react-query";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { ApiError } from "@/lib/api/types";
import { formatCurrencyVND } from "@/lib/utils";
import type { BookingDetail } from "@/types/booking.types";

interface BookingConfirmDialogProps {
  booking: BookingDetail;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  confirmMutation: UseMutationResult<unknown, Error, void>;
}

/**
 * FR-401 / BR-002 / decision B3: **warn, but allow**.
 *
 * When the booking has no payment, or its payment has not completed, this
 * shows a prominent warning — and leaves the confirm button ENABLED. Cash
 * and offline bank_transfer settlement are real, so blocking here would
 * break them. Disabling the button (or expecting the server to answer 409)
 * is the most likely defect in this screen; the server deliberately does not
 * enforce payment completion either.
 */
export function BookingConfirmDialog({ booking, open, onOpenChange, confirmMutation }: BookingConfirmDialogProps) {
  const unpaid = !booking.payment || booking.payment.status !== "completed";

  function handleConfirm() {
    confirmMutation.mutate(undefined, {
      onSuccess: () => {
        toast.success("Booking confirmed");
        onOpenChange(false);
      },
      onError: (err) => toast.error(err instanceof ApiError ? err.message : "Could not confirm this booking."),
    });
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Confirm booking"
      confirmLabel="Confirm booking"
      isPending={confirmMutation.isPending}
      onConfirm={handleConfirm}
      description={
        <div className="space-y-3">
          <p>
            Confirm <strong>{booking.booking_code}</strong> for <strong>{booking.contact_name}</strong> —{" "}
            {booking.num_participants} traveller(s), {formatCurrencyVND(booking.total_price)}.
          </p>
          {unpaid && (
            <div className="flex gap-2 rounded-md border border-amber-300 bg-amber-50 p-3 text-amber-900" role="alert">
              <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" aria-hidden />
              <p className="text-sm">
                {booking.payment
                  ? `This booking's payment is ${booking.payment.status}, not completed.`
                  : "No payment has been recorded for this booking."}{" "}
                You can still confirm it — settle cash or bank transfers outside the system as usual.
              </p>
            </div>
          )}
        </div>
      }
    />
  );
}

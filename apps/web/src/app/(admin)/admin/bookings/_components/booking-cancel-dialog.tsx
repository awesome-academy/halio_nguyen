"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import type { UseMutationResult } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api/types";
import { bookingCancelSchema, MAX_CANCELLATION_REASON, type BookingCancelFormValues } from "@/lib/validation/booking-schema";
import type { BookingCancelPayload, BookingDetail } from "@/types/booking.types";

interface BookingCancelDialogProps {
  booking: BookingDetail;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  cancelMutation: UseMutationResult<unknown, Error, BookingCancelPayload>;
}

/**
 * FR-402 / BR-004: cancelling requires a reason. The same cancel also
 * restores the schedule's seats (BR-005) — irreversibly, since no
 * compensating action exists — so the dialog says so plainly.
 */
export function BookingCancelDialog({ booking, open, onOpenChange, cancelMutation }: BookingCancelDialogProps) {
  const form = useForm<BookingCancelFormValues>({
    resolver: zodResolver(bookingCancelSchema),
    defaultValues: { cancellation_reason: "" },
  });

  useEffect(() => {
    if (open) form.reset({ cancellation_reason: "" });
  }, [open, form]);

  function handleSubmit(values: BookingCancelFormValues) {
    cancelMutation.mutate(values, {
      onSuccess: () => {
        toast.success("Booking cancelled");
        onOpenChange(false);
      },
      onError: (err) => {
        if (err instanceof ApiError && err.fields?.cancellation_reason) {
          form.setError("cancellation_reason", { type: "server", message: err.fields.cancellation_reason });
          return;
        }
        toast.error(err instanceof ApiError ? err.message : "Could not cancel this booking.");
      },
    });
  }

  const busy = cancelMutation.isPending;

  return (
    <Dialog open={open} onOpenChange={(o) => !busy && onOpenChange(o)}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Cancel booking</DialogTitle>
          <DialogDescription>
            {booking.booking_code} will be cancelled and its {booking.num_participants} seat(s) returned to the departure. This
            cannot be undone.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4" noValidate>
            <FormField
              control={form.control}
              name="cancellation_reason"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Reason</FormLabel>
                  <FormControl>
                    <Textarea
                      autoFocus
                      rows={4}
                      maxLength={MAX_CANCELLATION_REASON}
                      placeholder="Why is this booking being cancelled?"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={busy}>
                Keep booking
              </Button>
              <Button type="submit" variant="destructive" disabled={busy}>
                {busy && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                Cancel booking
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

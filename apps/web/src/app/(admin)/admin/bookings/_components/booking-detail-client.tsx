"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/admin/page-header";
import { BookingStatusBadge } from "@/components/admin/booking-status-badge";
import { getBooking } from "@/lib/api/bookings";
import { bookingKeys } from "@/lib/api/query-keys";
import { formatCurrencyVND, formatDate } from "@/lib/utils";
import type { BookingDetail } from "@/types/booking.types";
import { BookingPaymentCard } from "./booking-payment-card";
import { BookingStatusActions } from "./booking-status-actions";

interface BookingDetailClientProps {
  id: string;
}

/** SCR-BookingDetail (FR-211/FR-212). Read-only apart from the three SM-001
 * transitions — no field on a booking is editable here. */
export function BookingDetailClient({ id }: BookingDetailClientProps) {
  const { data: booking, isLoading, isError, refetch } = useQuery({
    queryKey: bookingKeys.detail(id),
    queryFn: () => getBooking(id),
  });

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-9 w-64" />
        <Skeleton className="h-48 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (isError || !booking) {
    return (
      <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
        <span>Could not load this booking.</span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Retry
          </Button>
          <Button asChild variant="ghost" size="sm">
            <Link href="/admin/bookings">Back to list</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <>
      <Button asChild variant="ghost" size="sm" className="mb-2 -ml-2">
        <Link href="/admin/bookings">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Booking Requests
        </Link>
      </Button>

      <PageHeader title={booking.booking_code} description={`Created ${formatDate(booking.created_at)}`}>
        <BookingStatusBadge status={booking.status} />
      </PageHeader>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <BookingSummary booking={booking} />
          <BookingContactPanel booking={booking} />
        </div>

        <div className="space-y-6">
          <BookingPaymentCard payment={booking.payment} />
          <section className="rounded-lg border p-5">
            <h2 className="text-sm font-semibold mb-3">Actions</h2>
            <BookingStatusActions booking={booking} />
          </section>
        </div>
      </div>
    </>
  );
}

function BookingSummary({ booking }: { booking: BookingDetail }) {
  return (
    <section className="rounded-lg border p-5">
      <h2 className="text-sm font-semibold mb-4">Tour &amp; departure</h2>
      <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
        <Field label="Tour" value={booking.tour.title} />
        <Field label="Destination" value={booking.tour.destination} />
        <Field label="Departs" value={formatDate(booking.schedule.departure_date)} />
        <Field label="Returns" value={formatDate(booking.schedule.return_date)} />
        <Field label="Travellers" value={String(booking.num_participants)} />
        <Field label="Unit price" value={formatCurrencyVND(booking.unit_price)} />
        <Field label="Total" value={formatCurrencyVND(booking.total_price)} />
        {booking.cancelled_at && <Field label="Cancelled" value={formatDate(booking.cancelled_at)} />}
      </dl>
      {booking.cancellation_reason && (
        <div className="mt-4 rounded-md bg-muted p-3">
          <p className="text-xs text-muted-foreground mb-1">Cancellation reason</p>
          {/* Free text from an admin: rendered as text, never as HTML. */}
          <p className="text-sm whitespace-pre-wrap break-words">{booking.cancellation_reason}</p>
        </div>
      )}
    </section>
  );
}

function BookingContactPanel({ booking }: { booking: BookingDetail }) {
  return (
    <section className="rounded-lg border p-5">
      <h2 className="text-sm font-semibold mb-4">Customer</h2>
      <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
        <Field label="Contact name" value={booking.contact_name} />
        <Field label="Account" value={booking.user.full_name} />
        <Field label="Phone" value={booking.contact_phone} />
        <Field label="Email" value={booking.contact_email} />
      </dl>
      {booking.special_requests && (
        <div className="mt-4 rounded-md bg-muted p-3">
          <p className="text-xs text-muted-foreground mb-1">Special requests</p>
          <p className="text-sm whitespace-pre-wrap break-words">{booking.special_requests}</p>
        </div>
      )}
    </section>
  );
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-0.5 break-words">{value}</dd>
    </div>
  );
}

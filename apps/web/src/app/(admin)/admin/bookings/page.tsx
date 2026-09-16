import { Suspense } from "react";
import { BookingsClient } from "./_components/bookings-client";

export const metadata = { title: "Booking Requests" };

/** SCR-BookingList — list state lives in the URL, so the client host needs a Suspense boundary for useSearchParams. */
export default function BookingsPage() {
  return (
    <Suspense>
      <BookingsClient />
    </Suspense>
  );
}

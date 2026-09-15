import { Suspense } from "react";
import { ToursClient } from "./_components/tours-client";

export const metadata = { title: "Tour Packages" };

/** SCR-TourList — list state lives in the URL, so the client host needs a Suspense boundary for useSearchParams. */
export default function ToursPage() {
  return (
    <Suspense>
      <ToursClient />
    </Suspense>
  );
}

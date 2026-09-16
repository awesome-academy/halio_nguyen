import { Suspense } from "react";
import { RevenueClient } from "./_components/revenue-client";

export const metadata = { title: "Revenue Analytics" };

/** SCR002 — range state lives in the URL, so the client host needs a Suspense boundary for useSearchParams. */
export default function AdminRevenuePage() {
  return (
    <Suspense>
      <RevenueClient />
    </Suspense>
  );
}

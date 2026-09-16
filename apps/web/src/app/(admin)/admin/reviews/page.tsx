import { Suspense } from "react";
import { ReviewsClient } from "./_components/reviews-client";

export const metadata = { title: "Reviews & Comments" };

/** SCR — list state lives in the URL, so the client host needs a Suspense boundary for useSearchParams. */
export default function ReviewsPage() {
  return (
    <Suspense>
      <ReviewsClient />
    </Suspense>
  );
}

import { Suspense } from "react";
import { CategoriesClient } from "./_components/categories-client";

export const metadata = { title: "Tour Categories" };

/** SCR-CategoryManagement — the list state lives in the URL, so the client host needs a Suspense boundary for useSearchParams. */
export default function CategoriesPage() {
  return (
    <Suspense>
      <CategoriesClient />
    </Suspense>
  );
}

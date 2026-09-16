import { Suspense } from "react";
import { UsersClient } from "./_components/users-client";

export const metadata = { title: "User Management" };

/** SCR001_UserListScreen — list state lives in the URL, so the client host needs a Suspense boundary for useSearchParams. */
export default function UsersPage() {
  return (
    <Suspense>
      <UsersClient />
    </Suspense>
  );
}

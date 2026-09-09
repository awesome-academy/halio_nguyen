"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";
import type { ApiError } from "@/lib/api/types";

function isUnauthorized(error: unknown): boolean {
  return typeof error === "object" && error !== null && (error as ApiError).status === 401;
}

/**
 * Mounts the TanStack Query provider. The QueryClient is created inside
 * useState (not at module scope) so a fresh cache is used per request on
 * the server and per mount on the client — a module-level client would leak
 * one request's cache into the next.
 */
export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            retry: (failureCount, error) => !isUnauthorized(error) && failureCount < 2,
            refetchOnWindowFocus: false,
          },
        },
      }),
  );

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

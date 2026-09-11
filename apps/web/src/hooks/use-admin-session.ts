"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { getSession, logout as logoutRequest, type AdminProfile } from "@/lib/api/auth";

export const adminSessionKey = ["admin", "session"] as const;

/**
 * Hydrates the real admin identity for the shell (FR-401) and doubles as
 * the expiry check (FR-402) — a 401 here fires apiFetch's onUnauthorized
 * callback without requiring a failed write action first.
 */
export function useAdminSession(options?: { enabled?: boolean }) {
  return useQuery<AdminProfile>({
    queryKey: adminSessionKey,
    queryFn: getSession,
    enabled: options?.enabled ?? true,
  });
}

/** Clears the query cache and returns to /admin/login regardless of outcome. */
export function useLogout() {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: logoutRequest,
    onSettled: () => {
      queryClient.clear();
      router.replace("/admin/login");
    },
  });
}

"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteUser, updateUser } from "@/lib/api/users";
import { userKeys } from "@/lib/api/query-keys";
import type { UserRole } from "@/types/user.types";

/**
 * One `useMutation` per FR-004/FR-005/FR-006 action. Unbound (each takes the
 * target `id` per call) so both the list's row actions (any row) and the
 * detail screen (its own id) share one mutation definition — same shape as
 * `useUpdateTourStatus`. Every mutation changes both the row in the list and
 * the detail page, so both are invalidated.
 */

export function useToggleUserStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, is_active }: { id: string; is_active: boolean }) => updateUser(id, { is_active }),
    onSuccess: (_data, { id }) => {
      void queryClient.invalidateQueries({ queryKey: userKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: userKeys.detail(id) });
    },
  });
}

export function useChangeUserRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: UserRole }) => updateUser(id, { role }),
    onSuccess: (_data, { id }) => {
      void queryClient.invalidateQueries({ queryKey: userKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: userKeys.detail(id) });
    },
  });
}

export function useDeleteUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: (_data, id) => {
      void queryClient.invalidateQueries({ queryKey: userKeys.lists() });
      // The detail id no longer exists in the active listing — drop it
      // rather than invalidate, so a stale detail view isn't refetched 404.
      queryClient.removeQueries({ queryKey: userKeys.detail(id) });
    },
  });
}

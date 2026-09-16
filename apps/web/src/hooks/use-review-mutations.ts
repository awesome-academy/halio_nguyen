"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteReview, updateReviewStatus } from "@/lib/api/reviews";
import { reviewKeys } from "@/lib/api/query-keys";
import type { ModerationStatus } from "@/types/review.types";

/**
 * A3/A4. Unbound (each takes the target `id` per call) so both the list's
 * row actions (any row) and the detail screen (its own id) share one
 * mutation definition — same shape as useToggleUserStatus/useDeleteUser.
 */

export function useUpdateReviewStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: ModerationStatus }) => updateReviewStatus(id, status),
    onSuccess: (_data, { id }) => {
      void queryClient.invalidateQueries({ queryKey: reviewKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: reviewKeys.detail(id) });
    },
  });
}

export function useDeleteReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteReview(id),
    onSuccess: (_data, id) => {
      void queryClient.invalidateQueries({ queryKey: reviewKeys.lists() });
      // Soft-deleted reviews vanish entirely (no restore) — drop the detail
      // query rather than invalidate, so a stale detail view isn't
      // refetched into a 404.
      queryClient.removeQueries({ queryKey: reviewKeys.detail(id) });
    },
  });
}

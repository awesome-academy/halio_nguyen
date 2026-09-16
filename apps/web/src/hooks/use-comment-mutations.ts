"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteComment, updateCommentVisibility } from "@/lib/api/reviews";
import { reviewKeys } from "@/lib/api/query-keys";

/**
 * A5/A6, scoped to one review. Hiding or deleting a comment changes that
 * review's trigger-maintained `comment_count` — a LIST column — so every
 * mutation here invalidates both the reviews list key and this review's
 * detail key. Invalidating only the detail key leaves a stale count sitting
 * in the table behind the detail page.
 */

export function useUpdateCommentVisibility(reviewId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ commentId, isHidden }: { commentId: string; isHidden: boolean }) =>
      updateCommentVisibility(reviewId, commentId, isHidden),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: reviewKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: reviewKeys.detail(reviewId) });
    },
  });
}

export function useDeleteComment(reviewId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (commentId: string) => deleteComment(reviewId, commentId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: reviewKeys.lists() });
      void queryClient.invalidateQueries({ queryKey: reviewKeys.detail(reviewId) });
    },
  });
}

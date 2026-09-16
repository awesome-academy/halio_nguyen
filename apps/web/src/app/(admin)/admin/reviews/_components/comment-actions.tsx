"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { useDeleteComment, useUpdateCommentVisibility } from "@/hooks/use-comment-mutations";
import { ApiError } from "@/lib/api/types";
import type { CommentNode } from "@/types/review.types";

interface CommentActionsProps {
  reviewId: string;
  comment: CommentNode;
}

/**
 * Hide/unhide + delete for one live comment (US006/US007). Renders nothing
 * for a deleted placeholder — a soft-deleted comment carries no content and
 * has nothing left to moderate.
 */
export function CommentActions({ reviewId, comment }: CommentActionsProps) {
  const [deleteOpen, setDeleteOpen] = useState(false);
  const updateVisibility = useUpdateCommentVisibility(reviewId);
  const remove = useDeleteComment(reviewId);

  if (comment.is_deleted) return null;

  const onError = (err: unknown) => toast.error(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");

  function toggleVisibility() {
    updateVisibility.mutate(
      { commentId: comment.id, isHidden: !comment.is_hidden },
      {
        onSuccess: () => toast.success(comment.is_hidden ? "Comment unhidden" : "Comment hidden"),
        onError,
      },
    );
  }

  function confirmDelete() {
    remove.mutate(comment.id, {
      onSuccess: () => {
        toast.success("Comment deleted");
        setDeleteOpen(false);
      },
      onError,
    });
  }

  return (
    <>
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" className="h-7 px-2 text-xs" disabled={updateVisibility.isPending} onClick={toggleVisibility}>
          {comment.is_hidden ? "Unhide" : "Hide"}
        </Button>
        <Button variant="ghost" size="sm" className="h-7 px-2 text-xs text-destructive" onClick={() => setDeleteOpen(true)}>
          Delete
        </Button>
      </div>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete comment"
        destructive
        confirmLabel="Delete"
        isPending={remove.isPending}
        onConfirm={confirmDelete}
        description={
          <p>This permanently removes the comment. Its replies remain and stay independently moderable. There is no restore.</p>
        }
      />
    </>
  );
}

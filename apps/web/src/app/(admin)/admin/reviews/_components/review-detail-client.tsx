"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/admin/page-header";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { ReviewStatusBadge } from "@/components/admin/review-status-badge";
import { useDeleteReview, useUpdateReviewStatus } from "@/hooks/use-review-mutations";
import { getReview } from "@/lib/api/reviews";
import { reviewKeys } from "@/lib/api/query-keys";
import { ApiError } from "@/lib/api/types";
import type { ModerationStatus } from "@/types/review.types";
import { ReviewContentPanel } from "./review-content-panel";
import { CommentThread } from "./comment-thread";

interface ReviewDetailClientProps {
  id: string;
}

/** SCR — review detail with the full comment thread (US002…US007). */
export function ReviewDetailClient({ id }: ReviewDetailClientProps) {
  const router = useRouter();
  const [statusTarget, setStatusTarget] = useState<ModerationStatus | null>(null);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const {
    data: review,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: reviewKeys.detail(id),
    queryFn: () => getReview(id),
    // A soft-deleted review 404s permanently — retrying it is pointless.
    retry: (failureCount, err) => !(err instanceof ApiError && err.status === 404) && failureCount < 3,
  });

  const updateStatus = useUpdateReviewStatus();
  const remove = useDeleteReview();

  const onError = (err: unknown) => toast.error(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-9 w-64" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  const notFound = error instanceof ApiError && error.status === 404;

  if (isError || !review) {
    return (
      <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
        <span>{notFound ? "This review has been deleted and is no longer available." : "Could not load this review."}</span>
        <div className="flex gap-2">
          {!notFound && (
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              Retry
            </Button>
          )}
          <Button asChild variant="ghost" size="sm">
            <Link href="/admin/reviews">Back to list</Link>
          </Button>
        </div>
      </div>
    );
  }

  function confirmToggle() {
    if (!statusTarget) return;
    updateStatus.mutate(
      { id, status: statusTarget },
      {
        onSuccess: () => {
          toast.success(statusTarget === "published" ? "Review published" : "Review hidden");
          setStatusTarget(null);
        },
        onError,
      },
    );
  }

  function handleDelete() {
    remove.mutate(id, {
      onSuccess: () => {
        toast.success("Review deleted");
        router.push("/admin/reviews");
      },
      onError,
    });
  }

  const toggleTo: ModerationStatus = review.status === "published" ? "hidden" : "published";

  return (
    <>
      <Button asChild variant="ghost" size="sm" className="mb-2 -ml-2">
        <Link href="/admin/reviews">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Reviews & Comments
        </Link>
      </Button>

      <PageHeader title={review.title} description={`by ${review.author_name}`}>
        <ReviewStatusBadge status={review.status} />
      </PageHeader>

      <div className="flex items-center gap-2 mb-6">
        <Button variant="outline" size="sm" onClick={() => setStatusTarget(toggleTo)}>
          {toggleTo === "hidden" ? "Hide" : "Publish"}
        </Button>
        <Button variant="destructive" size="sm" onClick={() => setDeleteOpen(true)}>
          Delete review
        </Button>
      </div>

      <div className="space-y-6">
        <ReviewContentPanel review={review} />
        <CommentThread reviewId={id} comments={review.comments} />
      </div>

      <ConfirmDialog
        open={statusTarget !== null}
        onOpenChange={(o) => !o && setStatusTarget(null)}
        title={statusTarget === "hidden" ? "Hide review" : "Publish review"}
        destructive={statusTarget === "hidden"}
        confirmLabel={statusTarget === "hidden" ? "Hide" : "Publish"}
        isPending={updateStatus.isPending}
        onConfirm={confirmToggle}
        description={
          <p>{statusTarget === "hidden" ? "This removes the review from public view." : "This makes the review visible to the public."}</p>
        }
      />

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete review"
        destructive
        confirmLabel="Delete"
        isPending={remove.isPending}
        onConfirm={handleDelete}
        description={<p>This permanently removes the review from the admin console. There is no restore.</p>}
      />
    </>
  );
}

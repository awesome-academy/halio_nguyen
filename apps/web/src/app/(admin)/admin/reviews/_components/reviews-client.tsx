"use client";

import { useMemo, useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/admin/page-header";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { DataTable } from "@/components/admin/data-table/data-table";
import { DataTableToolbar } from "@/components/admin/data-table/data-table-toolbar";
import { useListQueryState } from "@/hooks/use-list-query-state";
import { useDeleteReview, useUpdateReviewStatus } from "@/hooks/use-review-mutations";
import { listReviewCategories, listReviews, type ReviewListQuery } from "@/lib/api/reviews";
import { reviewCategoryKeys, reviewKeys } from "@/lib/api/query-keys";
import { ApiError } from "@/lib/api/types";
import type { ModerationStatus, ReviewListItem, ReviewStatus } from "@/types/review.types";
import { buildReviewColumns } from "./reviews-columns";
import { ReviewsFilterBar } from "./reviews-filter-bar";

const FILTER_KEYS = ["status", "review_category_id"];
// A7 is a seeded lookup, not a moving list — a long staleTime avoids
// refetching it on every list-page navigation.
const CATEGORIES_STALE_TIME_MS = 5 * 60 * 1000;

/** SCR — review list + moderation (US001, US003, US004, US005). */
export function ReviewsClient() {
  const list = useListQueryState({ filterKeys: FILTER_KEYS });
  const [statusTarget, setStatusTarget] = useState<{ review: ReviewListItem; target: ModerationStatus } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ReviewListItem | null>(null);

  const status = list.filters.status as ReviewStatus | undefined;
  const categoryId = list.filters.review_category_id;
  const apiQuery: ReviewListQuery = { ...list.query, status, review_category_id: categoryId };

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: reviewKeys.list(apiQuery),
    queryFn: () => listReviews(apiQuery),
    placeholderData: keepPreviousData,
  });
  const items = data?.items ?? [];
  const total = data?.total ?? 0;

  const { data: categories = [] } = useQuery({
    queryKey: reviewCategoryKeys.all,
    queryFn: listReviewCategories,
    staleTime: CATEGORIES_STALE_TIME_MS,
  });

  const updateStatus = useUpdateReviewStatus();
  const remove = useDeleteReview();

  const onError = (err: unknown) => toast.error(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");

  const columns = useMemo(
    () =>
      buildReviewColumns({
        onToggleStatus: (review, target) => setStatusTarget({ review, target }),
        onDelete: setDeleteTarget,
      }),
    [],
  );

  function confirmToggle() {
    if (!statusTarget) return;
    const { review, target } = statusTarget;
    updateStatus.mutate(
      { id: review.id, status: target },
      {
        onSuccess: () => {
          toast.success(target === "published" ? "Review published" : "Review hidden");
          setStatusTarget(null);
        },
        onError,
      },
    );
  }

  function confirmDelete() {
    if (!deleteTarget) return;
    remove.mutate(deleteTarget.id, {
      onSuccess: () => {
        toast.success("Review deleted");
        setDeleteTarget(null);
      },
      onError,
    });
  }

  return (
    <>
      <PageHeader title="Reviews & Comments" description="Moderate published reviews and their comment threads." />

      <DataTableToolbar search={list.search} onSearchChange={(search) => list.set({ search })} placeholder="Search by title…">
        <ReviewsFilterBar
          status={status}
          categoryId={categoryId}
          categories={categories}
          onStatusChange={(v) => list.set({ filters: { status: v } })}
          onCategoryChange={(v) => list.set({ filters: { review_category_id: v } })}
        />
      </DataTableToolbar>

      {isError ? (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
          <span>Could not load reviews.</span>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <DataTable
          columns={columns}
          data={items}
          getRowId={(r) => r.id}
          isLoading={isLoading}
          emptyMessage="No reviews match your search."
          page={list.page}
          pageSize={list.pageSize}
          total={total}
          sortBy={list.sortBy}
          sortDir={list.sortDir}
          onPageChange={(page) => list.set({ page })}
          onPageSizeChange={(pageSize) => list.set({ pageSize })}
          onSortChange={({ sortBy, sortDir }) => list.set({ sortBy, sortDir })}
        />
      )}

      <ConfirmDialog
        open={statusTarget !== null}
        onOpenChange={(o) => !o && setStatusTarget(null)}
        title={statusTarget?.target === "hidden" ? "Hide review" : "Publish review"}
        destructive={statusTarget?.target === "hidden"}
        confirmLabel={statusTarget?.target === "hidden" ? "Hide" : "Publish"}
        isPending={updateStatus.isPending}
        onConfirm={confirmToggle}
        description={
          statusTarget && (
            <p>
              {statusTarget.target === "hidden" ? "Hiding" : "Publishing"} <strong>{statusTarget.review.title}</strong>{" "}
              {statusTarget.target === "hidden" ? "removes it from public view." : "makes it visible to the public."}
            </p>
          )
        }
      />

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => !o && setDeleteTarget(null)}
        title="Delete review"
        destructive
        confirmLabel="Delete"
        isPending={remove.isPending}
        onConfirm={confirmDelete}
        description={
          deleteTarget && (
            <p>
              This permanently removes <strong>{deleteTarget.title}</strong> from the admin console. There is no restore.
            </p>
          )
        }
      />
    </>
  );
}

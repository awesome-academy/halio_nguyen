"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import type { UseMutationResult } from "@tanstack/react-query";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { ApiError } from "@/lib/api/types";
import type { TourListItem } from "@/types/tour.types";

interface TourDeleteDialogProps {
  tour: TourListItem | null;
  onOpenChange: (open: boolean) => void;
  onDeleted: () => void;
  deleteMutation: UseMutationResult<void, unknown, string>;
}

/**
 * D4: a normal destructive confirm, unless the server answers 409 — then it
 * re-renders as a blocking panel naming the active-booking count with no
 * force-delete path (the delete mutation is simply never retried).
 */
export function TourDeleteDialog({ tour, onOpenChange, onDeleted, deleteMutation }: TourDeleteDialogProps) {
  const [blockedCount, setBlockedCount] = useState<number | null>(null);

  useEffect(() => {
    if (tour === null) setBlockedCount(null);
  }, [tour]);

  function handleConfirm() {
    if (blockedCount !== null || !tour) {
      onOpenChange(false);
      return;
    }
    deleteMutation.mutate(tour.id, {
      onSuccess: () => {
        toast.success("Tour deleted");
        onDeleted();
      },
      onError: (err) => {
        if (err instanceof ApiError && err.status === 409) {
          const count = Number(err.fields?.booking_count ?? 0);
          setBlockedCount(count);
          return;
        }
        toast.error(err instanceof ApiError ? err.message : "Could not delete this tour.");
      },
    });
  }

  return (
    <ConfirmDialog
      open={tour !== null}
      onOpenChange={(o) => {
        if (!o) setBlockedCount(null);
        onOpenChange(o);
      }}
      title={blockedCount !== null ? "Cannot delete tour" : "Delete tour"}
      destructive={blockedCount === null}
      confirmLabel={blockedCount !== null ? "Close" : "Delete"}
      isPending={deleteMutation.isPending}
      onConfirm={handleConfirm}
      description={
        tour && (
          <>
            {blockedCount !== null ? (
              <p>
                <strong>{tour.title}</strong> cannot be deleted while <strong>{blockedCount}</strong> active booking(s) still reference it.
              </p>
            ) : (
              <p>
                This will remove <strong>{tour.title}</strong> from the catalogue. Its schedules and gallery are removed with it.
              </p>
            )}
          </>
        )
      }
    />
  );
}

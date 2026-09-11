"use client";

import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import type { CategoryListItem } from "@/types/category.types";

interface CategoryDeleteDialogProps {
  category: CategoryListItem | null;
  onOpenChange: (open: boolean) => void;
  onConfirm: (category: CategoryListItem) => void;
  isPending: boolean;
}

/** BR-001 in the UI: the confirm is disabled while tours are attached; the server's 409 remains the real guard. */
export function CategoryDeleteDialog({ category, onOpenChange, onConfirm, isPending }: CategoryDeleteDialogProps) {
  const blocked = (category?.tour_count ?? 0) > 0;
  return (
    <ConfirmDialog
      open={category !== null}
      onOpenChange={onOpenChange}
      title="Delete category"
      destructive
      confirmLabel="Delete"
      confirmDisabled={blocked}
      isPending={isPending}
      onConfirm={() => category && onConfirm(category)}
      description={
        category && (
          <>
            {blocked ? (
              <p>
                <strong>{category.name}</strong> cannot be deleted while <strong>{category.tour_count}</strong> tour(s) still use it. Reassign or
                delete those tours first.
              </p>
            ) : (
              <p>
                This will remove <strong>{category.name}</strong> from the catalogue. Its name and slug stay reserved and cannot be reused.
              </p>
            )}
          </>
        )
      }
    />
  );
}

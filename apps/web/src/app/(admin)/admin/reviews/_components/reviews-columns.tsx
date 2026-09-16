"use client";

import Link from "next/link";
import { createColumnHelper } from "@tanstack/react-table";
import { CheckCircle2, EyeOff, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ReviewStatusBadge } from "@/components/admin/review-status-badge";
import { DataTableColumnHeader } from "@/components/admin/data-table/data-table-column-header";
import type { DataTableColumnDef, DataTableFeatures } from "@/components/admin/data-table/types";
import { formatDate } from "@/lib/utils";
import type { ModerationStatus, ReviewListItem, ReviewStatus } from "@/types/review.types";

export interface ReviewRowActions {
  onToggleStatus: (review: ReviewListItem, target: ModerationStatus) => void;
  onDelete: (review: ReviewListItem) => void;
}

const helper = createColumnHelper<DataTableFeatures, ReviewListItem>();

/** SM-001: only "published" ever offers "Hide"; draft/hidden both offer "Publish" — draft is never targeted with "hidden" from here. */
function toggleTarget(status: ReviewStatus): { label: string; target: ModerationStatus } {
  return status === "published" ? { label: "Hide", target: "hidden" } : { label: "Publish", target: "published" };
}

/** FR-001's column set. Sortable columns are exactly the server's allowlist: created_at, like_count, comment_count, title. */
export function buildReviewColumns(actions: ReviewRowActions): DataTableColumnDef<ReviewListItem>[] {
  return [
    helper.accessor("title", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Title" />,
      cell: ({ row }) => (
        <Button asChild variant="link" className="h-auto p-0 font-medium">
          <Link href={`/admin/reviews/${row.original.id}`}>{row.original.title}</Link>
        </Button>
      ),
    }),
    helper.accessor("author_name", {
      header: "Author",
      enableSorting: false,
    }),
    helper.accessor("category_name", {
      header: "Category",
      enableSorting: false,
    }),
    helper.accessor("status", {
      header: "Status",
      enableSorting: false,
      cell: ({ getValue }) => <ReviewStatusBadge status={getValue()} />,
    }),
    helper.accessor("like_count", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Likes" />,
    }),
    helper.accessor("comment_count", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Comments" />,
    }),
    helper.accessor("created_at", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Created" />,
      cell: ({ getValue }) => <span className="text-sm text-muted-foreground">{formatDate(getValue())}</span>,
    }),
    helper.display({
      id: "actions",
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => {
        const r = row.original;
        const { label, target } = toggleTarget(r.status);
        const Icon = target === "hidden" ? EyeOff : CheckCircle2;
        return (
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              aria-label={`${label} ${r.title}`}
              onClick={() => actions.onToggleStatus(r, target)}
            >
              <Icon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-destructive"
              aria-label={`Delete ${r.title}`}
              onClick={() => actions.onDelete(r)}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        );
      },
    }),
  ];
}

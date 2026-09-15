"use client";

import { createColumnHelper } from "@tanstack/react-table";
import { Badge } from "@/components/ui/badge";
import { DataTableColumnHeader } from "@/components/admin/data-table/data-table-column-header";
import type { DataTableColumnDef, DataTableFeatures } from "@/components/admin/data-table/types";
import { formatCurrencyVND, formatDate } from "@/lib/utils";
import type { TourListItem, TourStatus } from "@/types/tour.types";
import { TourRowActions, type TourRowActionHandlers } from "./tour-row-actions";

const STATUS_VARIANT: Record<TourStatus, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Draft", variant: "outline" },
  published: { label: "Published", variant: "default" },
  archived: { label: "Archived", variant: "secondary" },
};

const helper = createColumnHelper<DataTableFeatures, TourListItem>();

interface BuildTourColumnsArgs extends TourRowActionHandlers {
  /** category_id -> name; List never joins category (see tours-client.tsx). */
  categoryNameById: Map<string, string>;
}

export function buildTourColumns({ categoryNameById, ...actions }: BuildTourColumnsArgs): DataTableColumnDef<TourListItem>[] {
  return [
    helper.accessor("title", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Title" />,
      cell: ({ row }) => (
        <div>
          <div className="font-medium">{row.original.title}</div>
          <div className="text-xs text-muted-foreground">{categoryNameById.get(row.original.category_id) ?? "—"}</div>
        </div>
      ),
    }),
    helper.accessor("destination", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Destination" />,
    }),
    helper.accessor("price", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Price" />,
      cell: ({ row }) => {
        const t = row.original;
        return t.discount_price !== undefined ? (
          <div>
            <div className="font-medium">{formatCurrencyVND(t.discount_price)}</div>
            <div className="text-xs text-muted-foreground line-through">{formatCurrencyVND(t.price)}</div>
          </div>
        ) : (
          <span>{formatCurrencyVND(t.price)}</span>
        );
      },
    }),
    helper.accessor("status", {
      header: "Status",
      enableSorting: false,
      cell: ({ row }) => {
        const s = STATUS_VARIANT[row.original.status];
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    }),
    helper.accessor("avg_rating", {
      header: "Rating",
      enableSorting: false,
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.avg_rating.toFixed(1)} ({row.original.total_ratings})
        </span>
      ),
    }),
    helper.accessor("created_at", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Created" />,
      cell: ({ getValue }) => <span className="text-sm text-muted-foreground">{formatDate(getValue())}</span>,
    }),
    helper.display({
      id: "actions",
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => <TourRowActions tour={row.original} {...actions} />,
    }),
  ];
}

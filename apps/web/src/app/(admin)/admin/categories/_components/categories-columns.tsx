"use client";

import { createColumnHelper } from "@tanstack/react-table";
import { ArrowDown, ArrowUp, Pencil, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTableColumnHeader } from "@/components/admin/data-table/data-table-column-header";
import type { DataTableColumnDef, DataTableFeatures } from "@/components/admin/data-table/types";
import { formatDate } from "@/lib/utils";
import type { CategoryListItem } from "@/types/category.types";

export interface CategoryRowActions {
  onEdit: (category: CategoryListItem) => void;
  onDelete: (category: CategoryListItem) => void;
  /** direction -1 = up, +1 = down; target ordinal is computed by the host. */
  onMove: (category: CategoryListItem, direction: -1 | 1) => void;
  /** Reorder only makes sense in the default display order with no filter. */
  canReorder: boolean;
  isFirst: (category: CategoryListItem) => boolean;
  isLast: (category: CategoryListItem) => boolean;
}

const helper = createColumnHelper<DataTableFeatures, CategoryListItem>();

export function buildCategoryColumns(actions: CategoryRowActions): DataTableColumnDef<CategoryListItem>[] {
  return [
    helper.accessor("sort_order", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Order" />,
      cell: ({ row }) => {
        const c = row.original;
        return (
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon" className="h-7 w-7" aria-label={`Move ${c.name} up`}
              disabled={!actions.canReorder || actions.isFirst(c)} onClick={() => actions.onMove(c, -1)}>
              <ArrowUp className="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" className="h-7 w-7" aria-label={`Move ${c.name} down`}
              disabled={!actions.canReorder || actions.isLast(c)} onClick={() => actions.onMove(c, 1)}>
              <ArrowDown className="h-3.5 w-3.5" />
            </Button>
          </div>
        );
      },
    }),
    helper.accessor("name", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
      cell: ({ row }) => (
        <div>
          <div className="font-medium">{row.original.name}</div>
          <div className="text-xs text-muted-foreground">/{row.original.slug}</div>
        </div>
      ),
    }),
    helper.accessor("tour_count", {
      header: "Tours",
      enableSorting: false,
      cell: ({ getValue }) => <Badge variant="secondary">{getValue()}</Badge>,
    }),
    helper.accessor("is_active", {
      header: "Status",
      enableSorting: false,
      cell: ({ getValue }) =>
        getValue() ? <Badge>Active</Badge> : <Badge variant="outline">Inactive</Badge>,
    }),
    helper.accessor("created_at", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Created" />,
      cell: ({ getValue }) => <span className="text-sm text-muted-foreground">{formatDate(getValue())}</span>,
    }),
    helper.display({
      id: "actions",
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => {
        const c = row.original;
        const blocked = c.tour_count > 0;
        return (
          <div className="flex items-center justify-end gap-1">
            <Button variant="ghost" size="icon" className="h-8 w-8" aria-label={`Edit ${c.name}`} onClick={() => actions.onEdit(c)}>
              <Pencil className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" aria-label={`Delete ${c.name}`}
              disabled={blocked}
              title={blocked ? `Cannot delete: ${c.tour_count} tour(s) still use this category` : undefined}
              onClick={() => actions.onDelete(c)}>
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        );
      },
    }),
  ];
}

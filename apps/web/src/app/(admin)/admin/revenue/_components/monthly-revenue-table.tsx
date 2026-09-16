"use client";

import { useMemo } from "react";
import { createColumnHelper } from "@tanstack/react-table";
import { DataTable } from "@/components/admin/data-table/data-table";
import { DataTableColumnHeader } from "@/components/admin/data-table/data-table-column-header";
import type {
  DataTableColumnDef,
  DataTableFeatures,
  DataTableSortState,
} from "@/components/admin/data-table/types";
import { formatCurrencyVND } from "@/lib/utils";
import type { MonthlyRevenueRow } from "@/types/revenue.types";

interface MonthlyRevenueTableProps extends DataTableSortState {
  rows: MonthlyRevenueRow[];
  isLoading: boolean;
  onSortChange: (sort: DataTableSortState) => void;
}

const helper = createColumnHelper<DataTableFeatures, MonthlyRevenueRow>();

/**
 * FR-002's columns. There is no tour column and there never can be: the
 * monthly view groups by month + category only (L6), so a per-tour monthly
 * breakdown does not exist to display.
 */
function buildColumns(): DataTableColumnDef<MonthlyRevenueRow>[] {
  return [
    helper.accessor("report_month", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Month" />,
      cell: ({ getValue }) => <span className="font-medium tabular-nums">{getValue()}</span>,
    }),
    helper.accessor("category_name", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Category" />,
    }),
    helper.accessor("total_bookings", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Bookings" />,
    }),
    helper.accessor("total_participants", {
      header: "Travellers",
      enableSorting: false,
    }),
    helper.accessor("total_revenue", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Revenue" />,
      cell: ({ getValue }) => <span className="font-semibold tabular-nums">{formatCurrencyVND(getValue())}</span>,
    }),
  ];
}

/**
 * Monthly revenue, one row per category per month. The whole range is
 * returned in one response (months x categories is small), so the table is
 * sorted but not paginated.
 */
export function MonthlyRevenueTable({ rows, isLoading, sortBy, sortDir, onSortChange }: MonthlyRevenueTableProps) {
  const columns = useMemo(buildColumns, []);

  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-lg font-semibold">Monthly revenue by category</h2>
        <p className="text-xs text-muted-foreground">
          One row per category per month. The monthly report has no per-tour breakdown.
        </p>
      </div>
      <DataTable
        columns={columns}
        data={rows}
        getRowId={(r) => `${r.report_month}-${r.category_id}`}
        isLoading={isLoading}
        emptyMessage="No revenue recorded for this period."
        page={1}
        pageSize={rows.length || 1}
        total={rows.length}
        sortBy={sortBy}
        sortDir={sortDir}
        onPageChange={() => undefined}
        onPageSizeChange={() => undefined}
        onSortChange={onSortChange}
      />
    </section>
  );
}

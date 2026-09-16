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
import { formatCurrencyVND, formatDate } from "@/lib/utils";
import type { DailyRevenueRow } from "@/types/revenue.types";

interface DailyRevenueTableProps extends DataTableSortState {
  rows: DailyRevenueRow[];
  total: number;
  page: number;
  pageSize: number;
  isLoading: boolean;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onSortChange: (sort: DataTableSortState) => void;
}

const helper = createColumnHelper<DataTableFeatures, DailyRevenueRow>();

/** FR-001's columns. Sortable ones are exactly the server's allowlist. */
function buildColumns(): DataTableColumnDef<DailyRevenueRow>[] {
  return [
    helper.accessor("report_date", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Date" />,
      cell: ({ getValue }) => <span className="font-medium">{formatDate(getValue())}</span>,
    }),
    helper.accessor("tour_title", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Tour Package" />,
    }),
    helper.accessor("category_name", {
      header: "Category",
      enableSorting: false,
      cell: ({ getValue }) => <span className="text-muted-foreground">{getValue()}</span>,
    }),
    helper.accessor("total_bookings", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Bookings" />,
    }),
    helper.accessor("total_participants", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Travellers" />,
    }),
    helper.accessor("total_revenue", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Revenue" />,
      cell: ({ getValue }) => <span className="font-semibold tabular-nums">{formatCurrencyVND(getValue())}</span>,
    }),
  ];
}

/** Daily revenue, one row per tour per day (mv_daily_revenue_report). */
export function DailyRevenueTable({ rows, ...state }: DailyRevenueTableProps) {
  const columns = useMemo(buildColumns, []);

  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-lg font-semibold">Daily revenue by tour</h2>
        <p className="text-xs text-muted-foreground">One row per tour per day within the selected range.</p>
      </div>
      <DataTable
        columns={columns}
        data={rows}
        getRowId={(r) => `${r.report_date}-${r.tour_id}`}
        emptyMessage="No revenue recorded for this period."
        {...state}
      />
    </section>
  );
}

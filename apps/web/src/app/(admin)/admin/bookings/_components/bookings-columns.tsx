"use client";

import Link from "next/link";
import { createColumnHelper } from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import { BookingStatusBadge } from "@/components/admin/booking-status-badge";
import { DataTableColumnHeader } from "@/components/admin/data-table/data-table-column-header";
import type { DataTableColumnDef, DataTableFeatures } from "@/components/admin/data-table/types";
import { formatCurrencyVND, formatDate } from "@/lib/utils";
import type { BookingListItem } from "@/types/booking.types";

const helper = createColumnHelper<DataTableFeatures, BookingListItem>();

/** FR-201's column set: booking code, customer, tour, pax, total, status,
 * created date. Sorting is server-side and only the three allowlisted
 * columns (created_at, total_price, status) may request it (FR-204). */
export function buildBookingColumns(): DataTableColumnDef<BookingListItem>[] {
  return [
    helper.accessor("booking_code", {
      header: "Booking",
      enableSorting: false,
      cell: ({ row }) => (
        <Button asChild variant="link" className="h-auto p-0 font-medium">
          <Link href={`/admin/bookings/${row.original.id}`}>{row.original.booking_code}</Link>
        </Button>
      ),
    }),
    helper.accessor("customer_name", {
      header: "Customer",
      enableSorting: false,
    }),
    helper.accessor("tour_title", {
      header: "Tour",
      enableSorting: false,
      cell: ({ row }) => (
        <div>
          <div className="font-medium">{row.original.tour_title}</div>
          <div className="text-xs text-muted-foreground">Departs {formatDate(row.original.departure_date)}</div>
        </div>
      ),
    }),
    helper.accessor("num_participants", {
      header: "Pax",
      enableSorting: false,
      cell: ({ getValue }) => <span className="tabular-nums">{getValue()}</span>,
    }),
    helper.accessor("total_price", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Total" />,
      cell: ({ getValue }) => <span className="font-medium tabular-nums">{formatCurrencyVND(getValue())}</span>,
    }),
    helper.accessor("status", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
      cell: ({ row }) => <BookingStatusBadge status={row.original.status} />,
    }),
    helper.accessor("created_at", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Created" />,
      cell: ({ getValue }) => <span className="text-sm text-muted-foreground">{formatDate(getValue())}</span>,
    }),
    helper.display({
      id: "actions",
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => (
        <Button asChild variant="ghost" size="sm">
          <Link href={`/admin/bookings/${row.original.id}`}>View</Link>
        </Button>
      ),
    }),
  ];
}

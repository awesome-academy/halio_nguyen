"use client";

import { useMemo } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/admin/page-header";
import { DataTable } from "@/components/admin/data-table/data-table";
import { DataTableToolbar } from "@/components/admin/data-table/data-table-toolbar";
import { useListQueryState } from "@/hooks/use-list-query-state";
import { listBookings, type BookingListQuery } from "@/lib/api/bookings";
import { listTours, type TourListQuery } from "@/lib/api/tours";
import { bookingKeys, tourKeys } from "@/lib/api/query-keys";
import type { BookingStatus } from "@/types/booking.types";
import { buildBookingColumns } from "./bookings-columns";
import { BookingsFilterBar } from "./bookings-filter-bar";

const FILTER_KEYS = ["status", "tour_id", "schedule_id", "date_from", "date_to"];

/** SCR-BookingList. Bookings are never created here — the customer site owns
 * creation — so this screen has no "New" action, only reads and drill-down. */
export function BookingsClient() {
  const list = useListQueryState({ filterKeys: FILTER_KEYS });

  const apiQuery: BookingListQuery = {
    ...list.query,
    status: list.filters.status,
    tour_id: list.filters.tour_id,
    schedule_id: list.filters.schedule_id,
    date_from: list.filters.date_from,
    date_to: list.filters.date_to,
  };

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: bookingKeys.list(apiQuery),
    queryFn: () => listBookings(apiQuery),
    placeholderData: keepPreviousData,
  });
  const items = data?.items ?? [];
  const total = data?.total ?? 0;

  // Tour select options for the filter bar. Published and draft tours can
  // both carry historical bookings, so this is not filtered by status.
  const tourQuery: TourListQuery = { page: 1, page_size: 100 };
  const { data: tourPage } = useQuery({
    queryKey: tourKeys.list(tourQuery),
    queryFn: () => listTours(tourQuery),
  });

  const columns = useMemo(() => buildBookingColumns(), []);

  return (
    <>
      <PageHeader title="Booking Requests" description="Review, confirm, complete and cancel customer bookings." />

      <DataTableToolbar
        search={list.search}
        onSearchChange={(search) => list.set({ search })}
        placeholder="Search by booking code, name or email…"
      >
        <BookingsFilterBar
          tours={tourPage?.items ?? []}
          status={list.filters.status as BookingStatus | undefined}
          tourId={list.filters.tour_id}
          scheduleId={list.filters.schedule_id}
          dateFrom={list.filters.date_from}
          dateTo={list.filters.date_to}
          onStatusChange={(v) => list.set({ filters: { status: v } })}
          onTourChange={(v) => list.set({ filters: { tour_id: v } })}
          onScheduleChange={(v) => list.set({ filters: { schedule_id: v } })}
          onDateChange={(range) => list.set({ filters: range })}
        />
      </DataTableToolbar>

      {isError ? (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
          <span>Could not load bookings.</span>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <DataTable
          columns={columns}
          data={items}
          getRowId={(b) => b.id}
          isLoading={isLoading}
          emptyMessage="No bookings found."
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
    </>
  );
}

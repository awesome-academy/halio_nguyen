"use client";

import { useQuery } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { getTour } from "@/lib/api/tours";
import { tourKeys } from "@/lib/api/query-keys";
import type { BookingStatus } from "@/types/booking.types";
import type { TourListItem } from "@/types/tour.types";

interface BookingsFilterBarProps {
  tours: TourListItem[];
  status: BookingStatus | undefined;
  tourId: string | undefined;
  scheduleId: string | undefined;
  dateFrom: string | undefined;
  dateTo: string | undefined;
  onStatusChange: (status: string | undefined) => void;
  onTourChange: (tourId: string | undefined) => void;
  onScheduleChange: (scheduleId: string | undefined) => void;
  onDateChange: (range: { date_from?: string; date_to?: string }) => void;
}

const STATUS_OPTIONS: { value: BookingStatus; label: string }[] = [
  { value: "pending", label: "Pending" },
  { value: "confirmed", label: "Confirmed" },
  { value: "completed", label: "Completed" },
  { value: "cancelled", label: "Cancelled" },
];

/** FR-202: status, tour, schedule and created-date range. The schedule select
 * is disabled until a tour is chosen — schedules only exist per tour, and the
 * API that lists them (A2) is keyed by tour id. */
export function BookingsFilterBar({
  tours,
  status,
  tourId,
  scheduleId,
  dateFrom,
  dateTo,
  onStatusChange,
  onTourChange,
  onScheduleChange,
  onDateChange,
}: BookingsFilterBarProps) {
  const { data: tourDetail } = useQuery({
    queryKey: tourKeys.detail(tourId ?? ""),
    queryFn: () => getTour(tourId as string),
    enabled: Boolean(tourId),
  });
  const schedules = tourDetail?.schedules ?? [];

  return (
    <>
      <Select value={status ?? "all"} onValueChange={(v) => onStatusChange(v === "all" ? undefined : v)}>
        <SelectTrigger className="w-40" aria-label="Status filter">
          <SelectValue placeholder="All statuses" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All statuses</SelectItem>
          {STATUS_OPTIONS.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={tourId ?? "all"}
        onValueChange={(v) => {
          // Changing the tour invalidates the chosen schedule: it belongs to
          // the previous tour and would return an empty list.
          onTourChange(v === "all" ? undefined : v);
          onScheduleChange(undefined);
        }}
      >
        <SelectTrigger className="w-56" aria-label="Tour filter">
          <SelectValue placeholder="All tours" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All tours</SelectItem>
          {tours.map((t) => (
            <SelectItem key={t.id} value={t.id}>
              {t.title}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Select
        value={scheduleId ?? "all"}
        onValueChange={(v) => onScheduleChange(v === "all" ? undefined : v)}
        disabled={!tourId}
      >
        <SelectTrigger className="w-52" aria-label="Departure filter">
          <SelectValue placeholder={tourId ? "All departures" : "Pick a tour first"} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All departures</SelectItem>
          {schedules.map((s) => (
            <SelectItem key={s.id} value={s.id}>
              {s.departure_date}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Input
        type="date"
        className="w-40"
        aria-label="Created from"
        value={dateFrom ?? ""}
        onChange={(e) => onDateChange({ date_from: e.target.value || undefined })}
      />
      <Input
        type="date"
        className="w-40"
        aria-label="Created to"
        value={dateTo ?? ""}
        onChange={(e) => onDateChange({ date_to: e.target.value || undefined })}
      />
    </>
  );
}

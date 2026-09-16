import { cn } from "@/lib/utils";
import type { BookingStatus } from "@/types/booking.types";

/**
 * The booking status palette, lifted out of the dashboard mock
 * (app/(admin)/admin/dashboard/page.tsx) so it survives that mock's removal
 * in phase 9. `cancelled` is the variant the mock never had.
 */
const STATUS_STYLES: Record<BookingStatus, { label: string; className: string }> = {
  pending: { label: "Pending", className: "text-amber-700 bg-amber-50 ring-amber-600/20" },
  confirmed: { label: "Confirmed", className: "text-emerald-700 bg-emerald-50 ring-emerald-600/20" },
  completed: { label: "Completed", className: "text-blue-700 bg-blue-50 ring-blue-600/20" },
  cancelled: { label: "Cancelled", className: "text-rose-700 bg-rose-50 ring-rose-600/20" },
};

interface BookingStatusBadgeProps {
  status: BookingStatus;
  className?: string;
}

export function BookingStatusBadge({ status, className }: BookingStatusBadgeProps) {
  const style = STATUS_STYLES[status];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[11px] font-semibold ring-1 ring-inset",
        style.className,
        className,
      )}
    >
      {style.label}
    </span>
  );
}

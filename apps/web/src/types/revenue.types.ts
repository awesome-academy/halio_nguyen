// Mirrors apps/api/internal/domain/revenue.go, which in turn mirrors the two
// materialized views. Every figure is batch data as of the last refresh —
// never live — which is why both screens carry that disclaimer.

/** One row of mv_daily_revenue_report: per day, per tour. */
export interface DailyRevenueRow {
  report_date: string;
  tour_id: string;
  tour_title: string;
  category_id: string;
  category_name: string;
  total_bookings: number;
  total_participants: number;
  total_revenue: number;
}

/**
 * One row of mv_monthly_revenue_report: per month, per category. There is
 * deliberately no tour here — the view groups by month + category only
 * (L6), and no monthly-by-tour breakdown exists to show.
 */
export interface MonthlyRevenueRow {
  report_month: string;
  category_id: string;
  category_name: string;
  total_bookings: number;
  total_participants: number;
  total_revenue: number;
}

/**
 * The shared list envelope plus last_refreshed_at. That field is null
 * whenever the API process has not completed a refresh since it started:
 * nothing persists it (DEC-001 / L3), so null is the honest answer and the
 * UI renders it as "unknown" rather than inventing a time.
 */
export interface RevenueListResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  last_refreshed_at: string | null;
}

/** A3's body: 202 when this call started a refresh, 409 when one is running. */
export interface RefreshRevenueResponse {
  status: "refresh_started" | "already_in_progress";
}

import { apiFetch } from "./client";
import { toQueryString } from "./query-string";
import type { ListQuery } from "./types";
import type {
  DailyRevenueRow,
  MonthlyRevenueRow,
  RefreshRevenueResponse,
  RevenueListResponse,
} from "@/types/revenue.types";

const BASE = "/api/v1/admin/revenue";

/** A1's query. from/to are ISO dates; both omitted means the current month. */
export interface DailyRevenueQuery extends ListQuery {
  from?: string;
  to?: string;
}

/**
 * A2's query. from/to are YYYY-MM — a malformed value is a 422, not an
 * empty table. There is no pagination: the whole range comes back.
 */
export interface MonthlyRevenueQuery {
  sort_by?: string;
  sort_dir?: "asc" | "desc";
  from?: string;
  to?: string;
}

// A1 — daily rows per tour for the range.
export function listDailyRevenue(query: DailyRevenueQuery): Promise<RevenueListResponse<DailyRevenueRow>> {
  return apiFetch<RevenueListResponse<DailyRevenueRow>>(`${BASE}/daily${toQueryString(query)}`);
}

// A2 — monthly rows per category for the range; also feeds the dashboard tile.
export function listMonthlyRevenue(query: MonthlyRevenueQuery): Promise<RevenueListResponse<MonthlyRevenueRow>> {
  return apiFetch<RevenueListResponse<MonthlyRevenueRow>>(`${BASE}/monthly${toQueryString(query)}`);
}

/**
 * A3 — starts a batch refresh of both views and returns immediately. A 409
 * (surfaced as an ApiError by apiFetch) means another refresh is already
 * running: it is rejected, never queued.
 */
export function refreshRevenue(): Promise<RefreshRevenueResponse> {
  return apiFetch<RefreshRevenueResponse>(`${BASE}/refresh`, { method: "POST" });
}

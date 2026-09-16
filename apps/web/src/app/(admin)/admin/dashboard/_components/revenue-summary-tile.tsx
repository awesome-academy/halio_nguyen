"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { ArrowUpRight, DollarSign } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { listMonthlyRevenue } from "@/lib/api/revenue";
import { revenueKeys } from "@/lib/api/query-keys";
import { formatCurrencyVND } from "@/lib/utils";
import type { MonthlyRevenueRow } from "@/types/revenue.types";

// Both params omitted means the current month server-side, which is exactly
// what this tile reports — so it needs no endpoint of its own.
const CURRENT_MONTH_QUERY = {};

function summarise(rows: MonthlyRevenueRow[]) {
  const total = rows.reduce((sum, row) => sum + row.total_revenue, 0);
  const top = rows.reduce<MonthlyRevenueRow | null>(
    (best, row) => (best === null || row.total_revenue > best.total_revenue ? row : best),
    null,
  );
  return { total, top };
}

/**
 * B1: the dashboard is revenue-only. One tile — this month's total plus the
 * top category — summed in the browser from A2's current-month rows, which
 * is a handful of category rows. A dedicated summary endpoint would be a
 * third route returning arithmetic the client can already do.
 */
export function RevenueSummaryTile() {
  const { data, isLoading, isError } = useQuery({
    queryKey: revenueKeys.monthly(CURRENT_MONTH_QUERY),
    queryFn: () => listMonthlyRevenue(CURRENT_MONTH_QUERY),
  });

  const { total, top } = summarise(data?.items ?? []);

  return (
    <div className="max-w-md space-y-3 rounded-xl border bg-card p-6 shadow-sm">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-muted-foreground">Revenue this month</span>
        <div className="rounded-lg bg-emerald-50 p-2 text-emerald-600">
          <DollarSign className="h-4 w-4" />
        </div>
      </div>

      {isLoading ? (
        <Skeleton className="h-9 w-48" />
      ) : isError ? (
        <p className="text-sm text-muted-foreground">Could not load revenue.</p>
      ) : (
        <>
          <div className="text-3xl font-bold tabular-nums">{formatCurrencyVND(total)}</div>
          <p className="text-xs text-muted-foreground">
            {top ? (
              <>
                Top category: <span className="font-medium text-foreground">{top.category_name}</span> (
                {formatCurrencyVND(top.total_revenue)})
              </>
            ) : (
              "No revenue recorded for this period."
            )}
          </p>
        </>
      )}

      {/* The number is batch data whose age this process may not know (L3). */}
      <p className="text-xs text-muted-foreground">As of the last batch refresh, not live.</p>

      <Link
        href="/admin/revenue"
        className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline"
      >
        Open revenue analytics
        <ArrowUpRight className="h-3 w-3" />
      </Link>
    </div>
  );
}

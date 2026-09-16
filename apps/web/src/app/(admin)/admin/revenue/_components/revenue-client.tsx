"use client";

import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/admin/page-header";
import { DateRangeFilter } from "@/components/admin/date-range-filter";
import { useListQueryState } from "@/hooks/use-list-query-state";
import { listDailyRevenue, listMonthlyRevenue } from "@/lib/api/revenue";
import { revenueKeys } from "@/lib/api/query-keys";
import { DailyRevenueTable } from "./daily-revenue-table";
import { MonthlyRevenueTable } from "./monthly-revenue-table";
import { RefreshRevenueButton } from "./refresh-revenue-button";
import { LastRefreshIndicator } from "./last-refresh-indicator";

// from/to are the shared range. The monthly table needs its own sort keys
// because the shared sort_by/sort_dir already belong to the daily table.
const FILTER_KEYS = ["from", "to", "m_sort_by", "m_sort_dir"];

/** SCR002 — daily and monthly revenue over one shared date range. */
export function RevenueClient() {
  const list = useListQueryState({ filterKeys: FILTER_KEYS });
  const { from = "", to = "" } = list.filters;

  const dailyQuery = { ...list.query, from: from || undefined, to: to || undefined };
  const monthlyQuery = {
    sort_by: list.filters.m_sort_by,
    sort_dir: list.filters.m_sort_dir as "asc" | "desc" | undefined,
    // report_month is YYYY-MM; the shared range is dates, so take the month
    // each endpoint of the range falls in.
    from: from ? from.slice(0, 7) : undefined,
    to: to ? to.slice(0, 7) : undefined,
  };

  const daily = useQuery({
    queryKey: revenueKeys.daily(dailyQuery),
    queryFn: () => listDailyRevenue(dailyQuery),
    placeholderData: keepPreviousData,
  });

  const monthly = useQuery({
    queryKey: revenueKeys.monthly(monthlyQuery),
    queryFn: () => listMonthlyRevenue(monthlyQuery),
    placeholderData: keepPreviousData,
  });

  // Either response carries it; they come from the same process.
  const lastRefreshedAt = daily.data?.last_refreshed_at ?? monthly.data?.last_refreshed_at ?? null;
  const isError = daily.isError || monthly.isError;

  return (
    <>
      <PageHeader title="Revenue Analytics" description="Batch revenue reports built from the payment ledger.">
        <RefreshRevenueButton />
      </PageHeader>

      <div className="mb-6 flex flex-wrap items-end justify-between gap-4 rounded-lg border bg-card p-4">
        <DateRangeFilter
          from={from}
          to={to}
          emptyHint="Showing the current month"
          onChange={(range) => list.set({ filters: range })}
        />
        <div className="flex flex-col items-end gap-1">
          <LastRefreshIndicator lastRefreshedAt={lastRefreshedAt} />
          {/* The figures are batch data, not live. Saying so next to the
              refresh control is what stops a stale number being quoted. */}
          <span className="text-xs text-muted-foreground">
            All figures are as of the last batch refresh, not live.
          </span>
        </div>
      </div>

      {isError ? (
        <div className="flex items-center justify-between rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm">
          <span>Could not load the revenue reports.</span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              void daily.refetch();
              void monthly.refetch();
            }}
          >
            Retry
          </Button>
        </div>
      ) : (
        <div className="space-y-8">
          <DailyRevenueTable
            rows={daily.data?.items ?? []}
            total={daily.data?.total ?? 0}
            page={list.page}
            pageSize={list.pageSize}
            isLoading={daily.isLoading}
            sortBy={list.sortBy}
            sortDir={list.sortDir}
            onPageChange={(page) => list.set({ page })}
            onPageSizeChange={(pageSize) => list.set({ pageSize })}
            onSortChange={({ sortBy, sortDir }) => list.set({ sortBy, sortDir })}
          />

          <MonthlyRevenueTable
            rows={monthly.data?.items ?? []}
            isLoading={monthly.isLoading}
            sortBy={list.filters.m_sort_by}
            sortDir={list.filters.m_sort_dir as "asc" | "desc" | undefined}
            onSortChange={({ sortBy, sortDir }) =>
              list.set({ filters: { m_sort_by: sortBy, m_sort_dir: sortDir } })
            }
          />
        </div>
      )}
    </>
  );
}

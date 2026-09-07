import { 
  BarChart3, 
  Calendar, 
  RefreshCw, 
  Download, 
  TrendingUp, 
  Layers, 
  DollarSign 
} from "lucide-react";
import { formatCurrencyVND } from "@/lib/utils";

// Mock data matching mv_daily_revenue_report & mv_monthly_revenue_report
const SAMPLE_DAILY_REVENUE = [
  { date: "2026-09-04", tour: "Phu Quoc Tropical Island Discovery", category: "Island & Coastal", bookings: 4, pax: 8, revenue: 31920000 },
  { date: "2026-09-03", tour: "Misty Sapa & Fansipan Peak", category: "Mountain & Trekking", bookings: 3, pax: 6, revenue: 14940000 },
  { date: "2026-09-02", tour: "Phu Quoc Tropical Island Discovery", category: "Island & Coastal", bookings: 5, pax: 10, revenue: 39900000 },
  { date: "2026-09-01", tour: "Misty Sapa & Fansipan Peak", category: "Mountain & Trekking", bookings: 2, pax: 4, revenue: 9960000 },
];

const SAMPLE_MONTHLY_REVENUE = [
  { month: "2026-09", category: "Island & Coastal Tours", bookings: 28, pax: 54, revenue: 215460000 },
  { month: "2026-09", category: "Mountain & Trekking", bookings: 16, pax: 32, revenue: 79680000 },
  { month: "2026-08", category: "Island & Coastal Tours", bookings: 45, pax: 98, revenue: 391020000 },
];

export default function AdminRevenuePage() {
  return (
    <div className="space-y-8 max-w-7xl mx-auto">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-extrabold tracking-tight">Revenue & Analytics (Batch Reports)</h1>
          <p className="text-sm text-muted-foreground">
            Aggregated financial reporting derived from PostgreSQL Materialized Views.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button 
            type="button"
            className="flex items-center gap-2 text-xs font-semibold px-3.5 py-2 rounded-lg border bg-card hover:bg-muted transition-colors"
          >
            <Download className="h-3.5 w-3.5" />
            <span>Export CSV</span>
          </button>

          <button 
            type="button"
            className="flex items-center gap-2 text-xs font-semibold px-3.5 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90 transition-opacity"
            title="Refreshes mv_daily_revenue_report & mv_monthly_revenue_report concurrently"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>Trigger Batch Refresh</span>
          </button>
        </div>
      </div>

      {/* Summary Highlights */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="p-5 rounded-xl border bg-card shadow-sm space-y-2">
          <span className="text-xs font-medium text-muted-foreground">Gross Booked Revenue (Sept 2026)</span>
          <div className="text-2xl font-bold text-emerald-600">{formatCurrencyVND(295140000)}</div>
          <span className="text-xs text-muted-foreground">Across 44 completed transactions</span>
        </div>

        <div className="p-5 rounded-xl border bg-card shadow-sm space-y-2">
          <span className="text-xs font-medium text-muted-foreground">Top Grossing Category</span>
          <div className="text-xl font-bold">Island & Coastal Tours</div>
          <span className="text-xs text-muted-foreground">73% of monthly revenue</span>
        </div>

        <div className="p-5 rounded-xl border bg-card shadow-sm space-y-2">
          <span className="text-xs font-medium text-muted-foreground">Materialized View Status</span>
          <div className="text-base font-bold text-emerald-600 flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rounded-full bg-emerald-500 animate-pulse" />
            <span>Synced</span>
          </div>
          <span className="text-xs text-muted-foreground">Last batch refresh: 10 minutes ago</span>
        </div>
      </div>

      {/* Daily Revenue Breakdown Table */}
      <div className="rounded-xl border bg-card shadow-sm overflow-hidden p-6 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-bold">Daily Revenue by Tour Package</h2>
            <p className="text-xs text-muted-foreground">Direct feed from mv_daily_revenue_report</p>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="border-b text-xs uppercase text-muted-foreground bg-muted/30">
              <tr>
                <th className="py-3 px-4">Date</th>
                <th className="py-3 px-4">Tour Package</th>
                <th className="py-3 px-4">Category</th>
                <th className="py-3 px-4 text-center">Bookings</th>
                <th className="py-3 px-4 text-center">Travelers (Pax)</th>
                <th className="py-3 px-4 text-right">Net Revenue</th>
              </tr>
            </thead>
            <tbody className="divide-y text-xs">
              {SAMPLE_DAILY_REVENUE.map((row, idx) => (
                <tr key={idx} className="hover:bg-muted/40 transition-colors">
                  <td className="py-3 px-4 font-mono font-medium">{row.date}</td>
                  <td className="py-3 px-4 font-semibold">{row.tour}</td>
                  <td className="py-3 px-4 text-muted-foreground">{row.category}</td>
                  <td className="py-3 px-4 text-center font-medium">{row.bookings}</td>
                  <td className="py-3 px-4 text-center font-medium">{row.pax}</td>
                  <td className="py-3 px-4 text-right font-bold text-emerald-600">
                    {formatCurrencyVND(row.revenue)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

import Link from "next/link";
import { 
  DollarSign, 
  CalendarCheck, 
  Map, 
  Star, 
  ArrowUpRight, 
  RefreshCw, 
  Plus,
  Clock,
  CheckCircle2,
  AlertCircle
} from "lucide-react";
import { formatCurrencyVND } from "@/lib/utils";

export default function AdminDashboardPage() {
  const KPIS = [
    { title: "Net Revenue (MTD)", value: formatCurrencyVND(148500000), change: "+18.4% vs last month", icon: DollarSign, color: "text-emerald-600 bg-emerald-50" },
    { title: "Total Bookings", value: "48", change: "6 pending confirmation", icon: CalendarCheck, color: "text-blue-600 bg-blue-50" },
    { title: "Active Tour Packages", value: "12", change: "4 categories active", icon: Map, color: "text-amber-600 bg-amber-50" },
    { title: "Verified Reviews", value: "248", change: "4.85 avg rating", icon: Star, color: "text-purple-600 bg-purple-50" },
  ];

  const RECENT_BOOKINGS = [
    { id: "b1", code: "SBT-20260904-A109", customer: "Nguyen Van An", tour: "Phu Quoc Tropical Island Discovery", people: 2, total: 7980000, status: "confirmed", date: "Today, 09:20" },
    { id: "b2", code: "SBT-20260904-B831", customer: "Tran Thi Mai", tour: "Misty Sapa & Fansipan Peak", people: 4, total: 9960000, status: "pending", date: "Today, 08:45" },
    { id: "b3", code: "SBT-20260903-C440", customer: "Le Hoang Nam", tour: "Phu Quoc Tropical Island Discovery", people: 1, total: 3990000, status: "completed", date: "Yesterday" },
  ];

  return (
    <div className="space-y-8 max-w-7xl mx-auto">
      {/* Header & Quick Action Buttons */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-extrabold tracking-tight">System Overview</h1>
          <p className="text-sm text-muted-foreground">Monitor performance, bookings, and revenue metrics.</p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/admin/revenue"
            className="flex items-center gap-1.5 text-xs font-semibold px-3 py-2 rounded-lg border bg-card hover:bg-muted transition-colors"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>Revenue Reports</span>
          </Link>

          <Link
            href="/admin/tours/new"
            className="flex items-center gap-1.5 text-xs font-semibold px-3.5 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90 transition-opacity"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>Create New Tour</span>
          </Link>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {KPIS.map((kpi, idx) => {
          const Icon = kpi.icon;
          return (
            <div key={idx} className="p-5 rounded-xl border bg-card shadow-sm space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-muted-foreground">{kpi.title}</span>
                <div className={`p-2 rounded-lg ${kpi.color}`}>
                  <Icon className="h-4 w-4" />
                </div>
              </div>
              <div>
                <div className="text-2xl font-bold">{kpi.value}</div>
                <div className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
                  <ArrowUpRight className="h-3 w-3 text-emerald-600" />
                  <span>{kpi.change}</span>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Recent Bookings Table */}
      <div className="rounded-xl border bg-card shadow-sm overflow-hidden space-y-4 p-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-bold">Recent Booking Requests</h2>
            <p className="text-xs text-muted-foreground">Latest reservations requiring review or payment verification</p>
          </div>

          <Link href="/admin/bookings" className="text-xs font-semibold text-primary hover:underline">
            View All Bookings
          </Link>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="border-b text-xs uppercase text-muted-foreground bg-muted/30">
              <tr>
                <th className="py-3 px-4">Booking Code</th>
                <th className="py-3 px-4">Customer</th>
                <th className="py-3 px-4">Tour Package</th>
                <th className="py-3 px-4 text-center">Pax</th>
                <th className="py-3 px-4">Total Amount</th>
                <th className="py-3 px-4 text-center">Status</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y text-xs">
              {RECENT_BOOKINGS.map((booking) => (
                <tr key={booking.id} className="hover:bg-muted/40 transition-colors">
                  <td className="py-3.5 px-4 font-mono font-semibold text-primary">
                    {booking.code}
                  </td>
                  <td className="py-3.5 px-4 font-medium">
                    {booking.customer}
                  </td>
                  <td className="py-3.5 px-4 max-w-xs truncate text-muted-foreground">
                    {booking.tour}
                  </td>
                  <td className="py-3.5 px-4 text-center font-medium">
                    {booking.people}
                  </td>
                  <td className="py-3.5 px-4 font-bold text-foreground">
                    {formatCurrencyVND(booking.total)}
                  </td>
                  <td className="py-3.5 px-4 text-center">
                    {booking.status === 'confirmed' && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-2.5 py-0.5 rounded-full">
                        <CheckCircle2 className="h-3 w-3" />
                        Confirmed
                      </span>
                    )}
                    {booking.status === 'pending' && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-amber-700 bg-amber-50 px-2.5 py-0.5 rounded-full">
                        <Clock className="h-3 w-3" />
                        Pending
                      </span>
                    )}
                    {booking.status === 'completed' && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-blue-700 bg-blue-50 px-2.5 py-0.5 rounded-full">
                        Completed
                      </span>
                    )}
                  </td>
                  <td className="py-3.5 px-4 text-right">
                    <Link
                      href={`/admin/bookings/${booking.id}`}
                      className="text-primary hover:underline font-semibold"
                    >
                      Manage
                    </Link>
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

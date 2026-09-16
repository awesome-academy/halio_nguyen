import { PageHeader } from "@/components/admin/page-header";
import { RevenueSummaryTile } from "./_components/revenue-summary-tile";

export const metadata = { title: "Dashboard" };

/**
 * SCR001 — revenue-only (B1). The cross-module KPI tiles and the recent
 * bookings table this page used to render were fabricated figures; they are
 * removed rather than rewired, because each module owns its own screen and a
 * made-up number sitting beside real ones invites the wrong one being trusted.
 */
export default function AdminDashboardPage() {
  return (
    <>
      <PageHeader title="Dashboard" description="Revenue at a glance. Each module's own screen carries its detail." />
      <RevenueSummaryTile />
    </>
  );
}

package repository

import "testing"

func TestDailyRevenueSortAllowlistNeverEchoesInput(t *testing.T) {
	// ORDER BY columns cannot be parameterized, so the allowlist's own
	// literal is the only string that may reach the SQL.
	col, dir := ResolveSort(dailyRevenueSortAllow, "total_revenue; DROP TABLE payments", "desc")
	if col != "report_date" {
		t.Errorf("hostile sort_by resolved to %q, want the default report_date", col)
	}
	if dir != SortDesc {
		t.Errorf("dir = %q, want desc", dir)
	}

	for by, want := range map[string]string{
		"report_date":        "report_date",
		"total_revenue":      "total_revenue",
		"total_bookings":     "total_bookings",
		"total_participants": "total_participants",
		"tour_title":         "tour_title",
		"":                   "report_date",
	} {
		if got, _ := ResolveSort(dailyRevenueSortAllow, by, ""); got != want {
			t.Errorf("sort_by=%q resolved to %q, want %q", by, got, want)
		}
	}
}

func TestMonthlyRevenueSortAllowlistNeverEchoesInput(t *testing.T) {
	col, dir := ResolveSort(monthlyRevenueSortAllow, "category_name UNION SELECT", "")
	if col != "report_month" {
		t.Errorf("hostile sort_by resolved to %q, want the default report_month", col)
	}
	if dir != SortAsc {
		t.Errorf("dir = %q, want asc", dir)
	}

	for by, want := range map[string]string{
		"report_month":   "report_month",
		"total_revenue":  "total_revenue",
		"total_bookings": "total_bookings",
		"category_name":  "category_name",
		"":               "report_month",
	} {
		if got, _ := ResolveSort(monthlyRevenueSortAllow, by, ""); got != want {
			t.Errorf("sort_by=%q resolved to %q, want %q", by, got, want)
		}
	}
}

func TestMonthlyRevenueHasNoTourSortColumn(t *testing.T) {
	// L6: mv_monthly_revenue_report groups by month + category only. A tour
	// sort would be an ORDER BY on a column the view does not have.
	for _, by := range []string{"tour_title", "tour_id"} {
		if col, _ := ResolveSort(monthlyRevenueSortAllow, by, ""); col != "report_month" {
			t.Errorf("sort_by=%q resolved to %q — the monthly view has no tour granularity", by, col)
		}
	}
}

func TestDailyRevenueParamsClampPaging(t *testing.T) {
	p := DailyRevenueParams{ListParams: ListParams{Page: 0, PageSize: 500}}
	p.Normalize()
	if p.Page != DefaultPage {
		t.Errorf("page = %d, want %d", p.Page, DefaultPage)
	}
	if p.PageSize != MaxPageSize {
		t.Errorf("page_size = %d, want it clamped to %d", p.PageSize, MaxPageSize)
	}

	p = DailyRevenueParams{ListParams: ListParams{Page: 3, PageSize: 20}}
	p.Normalize()
	if p.Offset() != 40 {
		t.Errorf("offset = %d, want 40", p.Offset())
	}
}

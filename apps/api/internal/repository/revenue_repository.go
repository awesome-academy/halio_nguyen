package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// DailyRevenueParams is A1's input: the shared page/sort fields plus the
// inclusive report_date range the service has already parsed and validated.
type DailyRevenueParams struct {
	ListParams
	From time.Time
	To   time.Time
}

// MonthlyRevenueParams is A2's input. It carries no pagination: the monthly
// view groups by month + category only (L6), so a range returns at most
// months x categories rows and is returned whole.
type MonthlyRevenueParams struct {
	SortBy  string
	SortDir string
	// From and To are YYYY-MM strings, compared lexicographically — correct
	// only because the service validates the zero-padded format first.
	From string
	To   string
}

// RevenueRepository reads the two materialized views F007 reports over. It
// declares no domain struct of its own: mv_daily_revenue_report and
// mv_monthly_revenue_report already map onto domain.DailyRevenueReport and
// domain.MonthlyRevenueReport.
type RevenueRepository interface {
	ListDaily(ctx context.Context, db DB, p DailyRevenueParams) ([]domain.DailyRevenueReport, int64, error)
	ListMonthly(ctx context.Context, db DB, p MonthlyRevenueParams) ([]domain.MonthlyRevenueReport, error)
}

// dailyRevenueSortAllow maps sort_by to the view's own columns; "" is
// FR-001's default (report_date). Direction is resolved separately.
var dailyRevenueSortAllow = map[string]string{
	"":                   "report_date",
	"report_date":        "report_date",
	"total_revenue":      "total_revenue",
	"total_bookings":     "total_bookings",
	"total_participants": "total_participants",
	"tour_title":         "tour_title",
}

// monthlyRevenueSortAllow is FR-002's allowlist. There is deliberately no
// tour column here — the monthly view has none (L6).
var monthlyRevenueSortAllow = map[string]string{
	"":               "report_month",
	"report_month":   "report_month",
	"total_revenue":  "total_revenue",
	"total_bookings": "total_bookings",
	"category_name":  "category_name",
}

const dailyRevenueColumns = `report_date, tour_id, tour_title, category_id, category_name, total_bookings, total_participants, total_revenue`

const monthlyRevenueColumns = `report_month, category_id, category_name, total_bookings, total_participants, total_revenue`

type revenueRepository struct{}

// NewRevenueRepository returns the pgx-backed RevenueRepository.
func NewRevenueRepository() RevenueRepository {
	return revenueRepository{}
}

func (revenueRepository) ListDaily(ctx context.Context, db DB, p DailyRevenueParams) ([]domain.DailyRevenueReport, int64, error) {
	p.Normalize()
	col, dir := ResolveSort(dailyRevenueSortAllow, p.SortBy, p.SortDir)

	rows, err := db.Query(ctx,
		`SELECT `+dailyRevenueColumns+`, COUNT(*) OVER() AS total_count
		 FROM mv_daily_revenue_report
		 WHERE report_date BETWEEN @from AND @to
		 ORDER BY `+col+` `+dir+`, tour_id
		 LIMIT @limit OFFSET @offset`,
		pgx.NamedArgs{"from": p.From, "to": p.To, "limit": p.PageSize, "offset": p.Offset()},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: listing daily revenue: %w", err)
	}
	defer rows.Close()

	items := make([]domain.DailyRevenueReport, 0, p.PageSize)
	var total int64
	for rows.Next() {
		var it domain.DailyRevenueReport
		if err := rows.Scan(
			&it.ReportDate, &it.TourID, &it.TourTitle, &it.CategoryID, &it.CategoryName,
			&it.TotalBookings, &it.TotalParticipants, &it.TotalRevenue, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scanning daily revenue: %w", err)
		}
		items = append(items, it)
	}
	return items, total, rows.Err()
}

func (revenueRepository) ListMonthly(ctx context.Context, db DB, p MonthlyRevenueParams) ([]domain.MonthlyRevenueReport, error) {
	col, dir := ResolveSort(monthlyRevenueSortAllow, p.SortBy, p.SortDir)

	rows, err := db.Query(ctx,
		`SELECT `+monthlyRevenueColumns+`
		 FROM mv_monthly_revenue_report
		 WHERE report_month BETWEEN @from AND @to
		 ORDER BY `+col+` `+dir+`, category_id`,
		pgx.NamedArgs{"from": p.From, "to": p.To},
	)
	if err != nil {
		return nil, fmt.Errorf("repository: listing monthly revenue: %w", err)
	}
	defer rows.Close()

	items := make([]domain.MonthlyRevenueReport, 0)
	for rows.Next() {
		var it domain.MonthlyRevenueReport
		if err := rows.Scan(
			&it.ReportMonth, &it.CategoryID, &it.CategoryName,
			&it.TotalBookings, &it.TotalParticipants, &it.TotalRevenue,
		); err != nil {
			return nil, fmt.Errorf("repository: scanning monthly revenue: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

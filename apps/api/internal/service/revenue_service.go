package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// isoDateLayout is A1's from/to format; monthLayout is A2's.
const (
	isoDateLayout = "2006-01-02"
	monthLayout   = "2006-01"
)

// errRefreshUnavailable is only reachable if the service was built without
// a runner, which router.Build never does.
var errRefreshUnavailable = errors.New("service: revenue refresh runner is not configured")

// monthPattern guards A2's lexicographic BETWEEN. report_month is a YYYY-MM
// *string*, so a string compare is only meaningful for that exact
// zero-padded shape — "2026-13" or "garbage" would otherwise return a
// silently empty table instead of an error.
var monthPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// RevenueService owns F007's read side: range validation, the current-month
// default, and the two view queries. The refresh trigger lives on the
// runner it holds (revenue_refresh_runner.go) — that is the only part of
// this feature with real failure modes.
type RevenueService struct {
	db     repository.DB
	repo   repository.RevenueRepository
	runner *RevenueRefreshRunner
}

// NewRevenueService wires the read path against the pool and the refresh
// runner. runner may be nil in tests that only exercise the reads.
func NewRevenueService(db repository.DB, repo repository.RevenueRepository, runner *RevenueRefreshRunner) *RevenueService {
	return &RevenueService{db: db, repo: repo, runner: runner}
}

// DailyRevenuePage is A1's result: the paginated rows plus the in-process
// last-refresh time (nil when this process has never run a refresh — L3).
type DailyRevenuePage struct {
	Items           []domain.DailyRevenueReport
	Total           int64
	Page            int
	PageSize        int
	LastRefreshedAt *time.Time
}

// MonthlyRevenuePage is A2's result. The monthly view is returned whole
// (months x categories rows), so the caller's total is simply len(Items).
type MonthlyRevenuePage struct {
	Items           []domain.MonthlyRevenueReport
	LastRefreshedAt *time.Time
}

// ListDaily is A1. from/to are ISO dates; both absent means the current
// month. The default sort is report_date DESC (newest first), which is what
// an admin opening the screen expects to see.
func (s *RevenueService) ListDaily(ctx context.Context, p repository.DailyRevenueParams, from, to string) (*DailyRevenuePage, error) {
	parsedFrom, parsedTo, err := resolveDateRange(from, to)
	if err != nil {
		return nil, err
	}
	p.From, p.To = parsedFrom, parsedTo
	p.Normalize()
	if p.SortDir == "" {
		p.SortDir = repository.SortDesc
	}

	items, total, err := s.repo.ListDaily(ctx, s.db, p)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &DailyRevenuePage{
		Items: items, Total: total, Page: p.Page, PageSize: p.PageSize,
		LastRefreshedAt: s.LastRefreshed(),
	}, nil
}

// ListMonthly is A2. from/to are YYYY-MM; both absent means the current
// month. Rows are category-level only — the view carries no tour (L6).
func (s *RevenueService) ListMonthly(ctx context.Context, sortBy, sortDir, from, to string) (*MonthlyRevenuePage, error) {
	parsedFrom, parsedTo, err := resolveMonthRange(from, to)
	if err != nil {
		return nil, err
	}
	if sortDir == "" {
		sortDir = repository.SortDesc
	}

	items, err := s.repo.ListMonthly(ctx, s.db, repository.MonthlyRevenueParams{
		SortBy: sortBy, SortDir: sortDir, From: parsedFrom, To: parsedTo,
	})
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &MonthlyRevenuePage{Items: items, LastRefreshedAt: s.LastRefreshed()}, nil
}

// TriggerRefresh is A3. It reports whether this call started a refresh;
// false means another one already holds the advisory lock (BR-001).
func (s *RevenueService) TriggerRefresh(ctx context.Context, actor string) (bool, error) {
	if s.runner == nil {
		return false, apperror.NewInternal(errRefreshUnavailable)
	}
	return s.runner.Trigger(ctx, actor)
}

// LastRefreshed returns the time this process last completed a refresh, or
// nil if it never has. nil is the correct answer after a restart, not a
// missing feature: no table stores this value and none may be added (L3).
func (s *RevenueService) LastRefreshed() *time.Time {
	if s.runner == nil {
		return nil
	}
	return s.runner.LastRefreshed()
}

// resolveDateRange parses A1's ISO from/to, defaulting each missing side to
// the current calendar month.
func resolveDateRange(from, to string) (time.Time, time.Time, error) {
	monthStart, monthEnd := currentMonthBounds(time.Now())

	parsedFrom, err := parseDateOr(from, monthStart, "from")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	parsedTo, err := parseDateOr(to, monthEnd, "to")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if parsedFrom.After(parsedTo) {
		return time.Time{}, time.Time{}, invalidRange()
	}
	return parsedFrom, parsedTo, nil
}

// resolveMonthRange validates A2's YYYY-MM from/to against monthPattern,
// defaulting each missing side to the current month.
func resolveMonthRange(from, to string) (string, string, error) {
	current := time.Now().Format(monthLayout)

	parsedFrom, err := parseMonthOr(from, current, "from")
	if err != nil {
		return "", "", err
	}
	parsedTo, err := parseMonthOr(to, current, "to")
	if err != nil {
		return "", "", err
	}
	if parsedFrom > parsedTo {
		return "", "", invalidRange()
	}
	return parsedFrom, parsedTo, nil
}

func parseDateOr(raw string, fallback time.Time, field string) (time.Time, error) {
	if raw == "" {
		return fallback, nil
	}
	parsed, err := time.Parse(isoDateLayout, raw)
	if err != nil {
		return time.Time{}, apperror.NewUnprocessable(
			"Date range is invalid",
			map[string]string{field: "Must be a date in YYYY-MM-DD format"},
		)
	}
	return parsed, nil
}

func parseMonthOr(raw, fallback, field string) (string, error) {
	if raw == "" {
		return fallback, nil
	}
	if !monthPattern.MatchString(raw) {
		return "", apperror.NewUnprocessable(
			"Month range is invalid",
			map[string]string{field: "Must be a month in YYYY-MM format"},
		)
	}
	return raw, nil
}

func invalidRange() error {
	return apperror.NewUnprocessable("Date range is invalid", map[string]string{"from": "Must not be after to"})
}

// currentMonthBounds returns the first and last day of now's month, in
// now's location, as the default A1 range.
func currentMonthBounds(now time.Time) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 1, -1)
}

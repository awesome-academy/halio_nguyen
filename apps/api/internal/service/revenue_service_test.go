package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// captureRevenueRepo records the params the service resolved so a test can
// assert the range and sort actually sent to the view query.
type captureRevenueRepo struct {
	daily   repository.DailyRevenueParams
	monthly repository.MonthlyRevenueParams
	called  bool
}

func (r *captureRevenueRepo) ListDaily(_ context.Context, _ repository.DB, p repository.DailyRevenueParams) ([]domain.DailyRevenueReport, int64, error) {
	r.daily, r.called = p, true
	return []domain.DailyRevenueReport{}, 0, nil
}

func (r *captureRevenueRepo) ListMonthly(_ context.Context, _ repository.DB, p repository.MonthlyRevenueParams) ([]domain.MonthlyRevenueReport, error) {
	r.monthly, r.called = p, true
	return []domain.MonthlyRevenueReport{}, nil
}

func newRevenueService(repo repository.RevenueRepository) *RevenueService {
	return NewRevenueService(nil, repo, nil)
}

// assertUnprocessable proves the service rejected the range as 422 rather
// than passing garbage to the query (which would return an empty table) or
// blowing up as a 500.
func assertUnprocessable(t *testing.T, err error, repo *captureRevenueRepo) {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("error %v is not an *apperror.Error", err)
	}
	if appErr.Code != apperror.CodeUnprocessable {
		t.Errorf("code = %v, want unprocessable (422)", appErr.Code)
	}
	if repo.called {
		t.Error("an invalid range must be rejected before the query runs")
	}
}

func TestListDailyRejectsMalformedDates(t *testing.T) {
	for _, tc := range []struct{ name, from, to string }{
		{"garbage from", "garbage", ""},
		{"garbage to", "", "not-a-date"},
		{"month only", "2026-09", ""},
		{"impossible date", "2026-02-30", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &captureRevenueRepo{}
			_, err := newRevenueService(repo).ListDaily(context.Background(), repository.DailyRevenueParams{}, tc.from, tc.to)
			assertUnprocessable(t, err, repo)
		})
	}
}

func TestListDailyRejectsInvertedRange(t *testing.T) {
	repo := &captureRevenueRepo{}
	_, err := newRevenueService(repo).ListDaily(context.Background(), repository.DailyRevenueParams{}, "2026-09-30", "2026-09-01")
	assertUnprocessable(t, err, repo)
}

func TestListDailyDefaultsToCurrentMonth(t *testing.T) {
	repo := &captureRevenueRepo{}
	if _, err := newRevenueService(repo).ListDaily(context.Background(), repository.DailyRevenueParams{}, "", ""); err != nil {
		t.Fatalf("ListDaily returned %v", err)
	}

	wantStart, wantEnd := currentMonthBounds(time.Now())
	if !repo.daily.From.Equal(wantStart) || !repo.daily.To.Equal(wantEnd) {
		t.Errorf("range = %v..%v, want the current month %v..%v", repo.daily.From, repo.daily.To, wantStart, wantEnd)
	}
	if repo.daily.SortDir != repository.SortDesc {
		t.Errorf("sort_dir = %q, want desc (newest first)", repo.daily.SortDir)
	}
	if repo.daily.PageSize != repository.DefaultPageSize {
		t.Errorf("page_size = %d, want %d", repo.daily.PageSize, repository.DefaultPageSize)
	}
}

func TestListDailyClampsOversizedPageSize(t *testing.T) {
	repo := &captureRevenueRepo{}
	p := repository.DailyRevenueParams{ListParams: repository.ListParams{PageSize: 5000}}
	if _, err := newRevenueService(repo).ListDaily(context.Background(), p, "", ""); err != nil {
		t.Fatalf("ListDaily returned %v", err)
	}
	if repo.daily.PageSize != repository.MaxPageSize {
		t.Errorf("page_size = %d, want it clamped to %d", repo.daily.PageSize, repository.MaxPageSize)
	}
}

func TestListMonthlyRejectsMalformedMonths(t *testing.T) {
	// report_month is a string compared lexicographically, so any of these
	// would otherwise return a silently empty table instead of an error.
	for _, tc := range []struct{ name, from, to string }{
		{"month 13", "2026-13", ""},
		{"month 00", "2026-00", ""},
		{"unpadded month", "2026-9", ""},
		{"garbage", "garbage", ""},
		{"full date", "2026-09-01", ""},
		{"garbage to", "", "2026-99"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &captureRevenueRepo{}
			_, err := newRevenueService(repo).ListMonthly(context.Background(), "", "", tc.from, tc.to)
			assertUnprocessable(t, err, repo)
		})
	}
}

func TestListMonthlyRejectsInvertedRange(t *testing.T) {
	repo := &captureRevenueRepo{}
	_, err := newRevenueService(repo).ListMonthly(context.Background(), "", "", "2026-09", "2026-08")
	assertUnprocessable(t, err, repo)
}

func TestListMonthlyDefaultsToCurrentMonth(t *testing.T) {
	repo := &captureRevenueRepo{}
	if _, err := newRevenueService(repo).ListMonthly(context.Background(), "", "", "", ""); err != nil {
		t.Fatalf("ListMonthly returned %v", err)
	}

	want := time.Now().Format(monthLayout)
	if repo.monthly.From != want || repo.monthly.To != want {
		t.Errorf("range = %q..%q, want the current month %q", repo.monthly.From, repo.monthly.To, want)
	}
	if repo.monthly.SortDir != repository.SortDesc {
		t.Errorf("sort_dir = %q, want desc", repo.monthly.SortDir)
	}
}

func TestListMonthlyAcceptsAValidRange(t *testing.T) {
	repo := &captureRevenueRepo{}
	if _, err := newRevenueService(repo).ListMonthly(context.Background(), "total_revenue", "asc", "2026-01", "2026-12"); err != nil {
		t.Fatalf("ListMonthly returned %v", err)
	}
	if repo.monthly.From != "2026-01" || repo.monthly.To != "2026-12" {
		t.Errorf("range = %q..%q, want 2026-01..2026-12", repo.monthly.From, repo.monthly.To)
	}
	if repo.monthly.SortDir != "asc" {
		t.Errorf("sort_dir = %q, want the caller's asc", repo.monthly.SortDir)
	}
}

func TestLastRefreshedIsNilWithoutARunner(t *testing.T) {
	// L3: no table stores this value, so a process that has never run a
	// refresh must answer nil — the UI turns that into "unknown".
	if got := newRevenueService(&captureRevenueRepo{}).LastRefreshed(); got != nil {
		t.Errorf("LastRefreshed = %v, want nil", got)
	}
}

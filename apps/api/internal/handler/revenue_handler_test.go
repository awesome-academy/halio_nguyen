package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

type stubRevenueRepo struct {
	daily   []domain.DailyRevenueReport
	monthly []domain.MonthlyRevenueReport
}

func (s stubRevenueRepo) ListDaily(context.Context, repository.DB, repository.DailyRevenueParams) ([]domain.DailyRevenueReport, int64, error) {
	return s.daily, int64(len(s.daily)), nil
}

func (s stubRevenueRepo) ListMonthly(context.Context, repository.DB, repository.MonthlyRevenueParams) ([]domain.MonthlyRevenueReport, error) {
	return s.monthly, nil
}

// lockRow answers pg_try_advisory_lock. It hands the lock to the first
// caller only, the way Postgres does.
type lockRow struct{ acquired bool }

func (r lockRow) Scan(dest ...any) error {
	if p, ok := dest[0].(*bool); ok {
		*p = r.acquired
	}
	return nil
}

// lockingConnSource models the advisory lock: the first Trigger wins, every
// later one is refused until the refresh releases it.
type lockingConnSource struct {
	mu      sync.Mutex
	held    bool
	release chan struct{}
}

type lockingConn struct {
	src      *lockingConnSource
	acquired bool
}

func (s *lockingConnSource) Acquire(context.Context) (service.RefreshConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acquired := !s.held
	if acquired {
		s.held = true
	}
	return &lockingConn{src: s, acquired: acquired}, nil
}

func (c *lockingConn) QueryRow(context.Context, string, ...any) pgx.Row {
	return lockRow{acquired: c.acquired}
}

func (c *lockingConn) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	if strings.HasPrefix(sql, "CALL") {
		<-c.src.release // hold the lock until the test lets go
		return pgconn.CommandTag{}, nil
	}
	c.src.mu.Lock()
	c.src.held = false
	c.src.mu.Unlock()
	return pgconn.CommandTag{}, nil
}

func (c *lockingConn) Release() {}

func revenueGate(t *testing.T, repo repository.RevenueRepository, conns service.ConnSource) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	gated := e.Group("/api/v1/admin")
	gated.Use(jwtGate(t))

	var runner *service.RevenueRefreshRunner
	if conns != nil {
		runner = service.NewRevenueRefreshRunner(conns)
	}
	NewRevenueHandler(service.NewRevenueService(nil, repo, runner)).RegisterRoutes(gated)
	return e
}

func revenueRoutes() []struct{ method, path string } {
	return []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/revenue/daily"},
		{http.MethodGet, "/api/v1/admin/revenue/monthly"},
		{http.MethodPost, "/api/v1/admin/revenue/refresh"},
	}
}

func TestAllRevenueRoutesRejectMissingCookie(t *testing.T) {
	e := revenueGate(t, stubRevenueRepo{}, nil)
	for _, r := range revenueRoutes() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without cookie: %d, want 401", r.method, r.path, rec.Code)
		}
	}
}

func TestAllRevenueRoutesRejectNonAdminRole(t *testing.T) {
	e := revenueGate(t, stubRevenueRepo{}, nil)
	token, _ := jwtutil.Issue(testSecret, time.Hour, "u1", "user")
	for _, r := range revenueRoutes() {
		req := httptest.NewRequest(r.method, r.path, nil)
		req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with role=user: %d, want 403", r.method, r.path, rec.Code)
		}
	}
}

func TestListDailyReturnsTheSharedEnvelopeWithLastRefreshed(t *testing.T) {
	e := revenueGate(t, stubRevenueRepo{daily: []domain.DailyRevenueReport{{TourTitle: "Sapa", TotalRevenue: 1500}}}, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, adminRequest(t, http.MethodGet, "/api/v1/admin/revenue/daily?sort_by=total_revenue;DROP&page_size=5", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	// last_refreshed_at is null, not absent and not invented (DEC-001/L3).
	for _, want := range []string{`"items"`, `"total":1`, `"page":1`, `"page_size":5`, `"last_refreshed_at":null`, `"tour_title":"Sapa"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("body missing %s: %s", want, rec.Body.String())
		}
	}
}

func TestListMonthlyReturnsCategoryRowsWholeRange(t *testing.T) {
	e := revenueGate(t, stubRevenueRepo{monthly: []domain.MonthlyRevenueReport{
		{ReportMonth: "2026-09", CategoryName: "Island"},
		{ReportMonth: "2026-08", CategoryName: "Mountain"},
	}}, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, adminRequest(t, http.MethodGet, "/api/v1/admin/revenue/monthly?from=2026-08&to=2026-09", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"total":2`) || !strings.Contains(body, `"category_name":"Island"`) {
		t.Errorf("unexpected body: %s", body)
	}
	if strings.Contains(body, "tour_title") || strings.Contains(body, "tour_id") {
		t.Errorf("monthly rows must carry no tour granularity (L6): %s", body)
	}
}

func TestMalformedMonthIs422NotAnEmptyTable(t *testing.T) {
	e := revenueGate(t, stubRevenueRepo{}, nil)
	for _, q := range []string{"from=2026-13", "from=garbage", "to=2026-9", "from=2026-09&to=2026-08"} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, adminRequest(t, http.MethodGet, "/api/v1/admin/revenue/monthly?"+q, ""))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("?%s: status %d, want 422 (body %s)", q, rec.Code, rec.Body.String())
		}
	}
}

func TestMalformedDailyDateIs422(t *testing.T) {
	e := revenueGate(t, stubRevenueRepo{}, nil)
	for _, q := range []string{"from=garbage", "to=2026-99-99", "from=2026-09-30&to=2026-09-01"} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, adminRequest(t, http.MethodGet, "/api/v1/admin/revenue/daily?"+q, ""))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("?%s: status %d, want 422 (body %s)", q, rec.Code, rec.Body.String())
		}
	}
}

func TestRefreshReturns202ThenConflictsWhileInFlight(t *testing.T) {
	src := &lockingConnSource{release: make(chan struct{})}
	e := revenueGate(t, stubRevenueRepo{}, src)

	first := httptest.NewRecorder()
	e.ServeHTTP(first, adminRequest(t, http.MethodPost, "/api/v1/admin/revenue/refresh", ""))
	if first.Code != http.StatusAccepted {
		t.Fatalf("first refresh: %d, want 202 (body %s)", first.Code, first.Body.String())
	}
	if !strings.Contains(first.Body.String(), `"refresh_started"`) {
		t.Errorf("first refresh body = %s", first.Body.String())
	}

	// The first refresh is still blocked inside the CALL, so the lock is
	// held and a second trigger must be rejected, never queued (BR-001).
	second := httptest.NewRecorder()
	e.ServeHTTP(second, adminRequest(t, http.MethodPost, "/api/v1/admin/revenue/refresh", ""))
	if second.Code != http.StatusConflict {
		t.Fatalf("second refresh: %d, want 409 (body %s)", second.Code, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), `"already_in_progress"`) {
		t.Errorf("second refresh body = %s", second.Body.String())
	}

	close(src.release)
}

func TestRefreshRespondsBeforeTheRefreshFinishes(t *testing.T) {
	src := &lockingConnSource{release: make(chan struct{})}
	e := revenueGate(t, stubRevenueRepo{}, src)

	// The CALL cannot complete until release is closed, so a handler that
	// waited for it would never return here.
	done := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, adminRequest(t, http.MethodPost, "/api/v1/admin/revenue/refresh", ""))
		done <- rec.Code
	}()

	select {
	case code := <-done:
		if code != http.StatusAccepted {
			t.Errorf("status %d, want 202", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("A3 blocked on the refresh instead of returning 202 immediately")
	}
	close(src.release)
}

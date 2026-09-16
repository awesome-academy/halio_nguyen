package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// F005's A0 gate plus the scaffolding user_admin_handler_security_test.go
// shares. Split to keep each file under the 200-line cap — same arrangement
// as the booking handler tests.

const (
	testUserAdminActorID  = "77777777-7777-7777-7777-777777777001"
	testUserAdminTargetID = "77777777-7777-7777-7777-777777777002"
)

// stubUserAdminRepo answers every UserAdminRepository method with fixtures.
// Unlike the read-only booking/category stubs, A3/A4 handler tests here
// exercise a real (mocked) transaction end to end, so every method needed by
// that path is implemented rather than inherited-and-panicking.
type stubUserAdminRepo struct {
	repository.UserAdminRepository
	items        []domain.User
	total        int64
	detail       *domain.User
	getErr       error
	bookingCount int64
	reviewCount  int64
	lockedAdmins []uuid.UUID
	updated      *domain.User
}

func (s *stubUserAdminRepo) List(context.Context, repository.DB, repository.UserListParams) ([]domain.User, int64, error) {
	return s.items, s.total, nil
}

func (s *stubUserAdminRepo) Get(context.Context, repository.DB, uuid.UUID) (*domain.User, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.detail, nil
}

func (s *stubUserAdminRepo) CountBookings(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return s.bookingCount, nil
}

func (s *stubUserAdminRepo) CountReviews(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return s.reviewCount, nil
}

func (s *stubUserAdminRepo) LockActiveAdmins(context.Context, repository.DB) ([]uuid.UUID, error) {
	return s.lockedAdmins, nil
}

func (s *stubUserAdminRepo) Update(_ context.Context, _ repository.DB, id uuid.UUID, patch repository.UserPatch) (*domain.User, error) {
	if s.updated != nil {
		return s.updated, nil
	}
	u := &domain.User{ID: id}
	if patch.Role != nil {
		u.Role = *patch.Role
	}
	if patch.IsActive != nil {
		u.IsActive = *patch.IsActive
	}
	return u, nil
}

func (s *stubUserAdminRepo) SoftDelete(context.Context, repository.DB, uuid.UUID) error {
	return nil
}

// stubActiveBookingRepo answers only CountActiveByUser (BR-003); any other
// call panics loudly via the embedded interface.
type stubActiveBookingRepo struct {
	repository.BookingRepository
	activeCount int64
}

func (s stubActiveBookingRepo) CountActiveByUser(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return s.activeCount, nil
}

// stubTx/stubUserAdminDB let A3/A4 open and commit a transaction without a
// real pool — the mocked repo methods above ignore the tx argument entirely,
// so no real SQL ever runs against it.
type stubTx struct{ pgx.Tx }

func (stubTx) Commit(context.Context) error   { return nil }
func (stubTx) Rollback(context.Context) error { return nil }

type stubUserAdminDB struct{ repository.DB }

func (stubUserAdminDB) Begin(context.Context) (pgx.Tx, error) { return stubTx{}, nil }

// userAdminGate wires the F005 handler behind the same A0 chain router.New
// installs, so the gate is proven by request rather than by inspection.
func userAdminGate(t *testing.T, repo repository.UserAdminRepository, bookingRepo repository.BookingRepository) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	gated := e.Group("/api/v1/admin")
	gated.Use(jwtGate(t))

	svc := service.NewUserAdminService(stubUserAdminDB{}, repo, bookingRepo)
	NewUserAdminHandler(svc).RegisterRoutes(gated)
	return e
}

func userAdminRequest(t *testing.T, method, path, body, subject, role string) *http.Request {
	t.Helper()
	token, err := jwtutil.Issue(testSecret, time.Hour, subject, role)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
	return req
}

// userAdminRoutes lists all 4 F005 actions against a fixed target id.
func userAdminRoutes(id string) []struct{ method, path string } {
	return []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/users"},
		{http.MethodGet, "/api/v1/admin/users/" + id},
		{http.MethodPatch, "/api/v1/admin/users/" + id},
		{http.MethodDelete, "/api/v1/admin/users/" + id},
	}
}

func TestAllUserAdminRoutesRejectMissingCookie(t *testing.T) {
	e := userAdminGate(t, &stubUserAdminRepo{}, stubActiveBookingRepo{})
	for _, r := range userAdminRoutes(testUserAdminTargetID) {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without cookie: %d, want 401", r.method, r.path, rec.Code)
		}
	}
}

func TestAllUserAdminRoutesRejectNonAdminRole(t *testing.T) {
	e := userAdminGate(t, &stubUserAdminRepo{}, stubActiveBookingRepo{})
	for _, r := range userAdminRoutes(testUserAdminTargetID) {
		req := userAdminRequest(t, r.method, r.path, "", testUserAdminActorID, "user")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with role=user: %d, want 403", r.method, r.path, rec.Code)
		}
	}
}

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// F004's A0 gate plus the scaffolding the other two booking handler test
// files share (params, security). Split to keep each file under the
// 200-line cap.

const (
	testBookingID = "44444444-4444-4444-4444-444444444001"
	// A real UUID, unlike the other handler tests' "u1": every A3-A5
	// transition resolves the acting admin from this claim (L1), so it has
	// to parse.
	testAdminSubject = "44444444-4444-4444-4444-4444444440ad"
)

// stubBookingReadRepo answers A1/A2 with fixtures. Every other method is
// inherited from the embedded interface and panics loudly if a test reaches
// it — these tests exercise the gate and the handler's own parsing, not the
// service's transition rules (those are booking_service_test.go's).
type stubBookingReadRepo struct {
	repository.BookingRepository
	items   []domain.BookingListItem
	detail  *domain.Booking
	lastPar repository.BookingListParams
}

func (s *stubBookingReadRepo) List(_ context.Context, _ repository.DB, p repository.BookingListParams) ([]domain.BookingListItem, int64, error) {
	s.lastPar = p
	return s.items, int64(len(s.items)), nil
}

func (s *stubBookingReadRepo) FindByID(context.Context, repository.DB, uuid.UUID) (*domain.Booking, error) {
	if s.detail == nil {
		return nil, repository.ErrBookingNotFound
	}
	return s.detail, nil
}

// bookingGate wires the F004 handler behind the same A0 chain router.New
// installs, so the gate is proven by request rather than by inspection.
func bookingGate(t *testing.T, repo repository.BookingRepository) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	gated := e.Group("/api/v1/admin")
	gated.Use(jwtGate(t))

	NewBookingHandler(service.NewBookingService(nil, repo, stubTourScheduleRepo{}, nil)).RegisterRoutes(gated)
	return e
}

func bookingRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	token, err := jwtutil.Issue(testSecret, time.Hour, testAdminSubject, "admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
	return req
}

// bookingRoutes lists all 5 F004 actions.
func bookingRoutes() []struct{ method, path string } {
	return []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/bookings"},
		{http.MethodGet, "/api/v1/admin/bookings/" + testBookingID},
		{http.MethodPatch, "/api/v1/admin/bookings/" + testBookingID + "/confirm"},
		{http.MethodPatch, "/api/v1/admin/bookings/" + testBookingID + "/cancel"},
		{http.MethodPatch, "/api/v1/admin/bookings/" + testBookingID + "/complete"},
	}
}

func TestAllBookingRoutesRejectMissingCookie(t *testing.T) {
	e := bookingGate(t, &stubBookingReadRepo{})
	for _, r := range bookingRoutes() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without cookie: %d, want 401", r.method, r.path, rec.Code)
		}
	}
}

func TestAllBookingRoutesRejectNonAdminRole(t *testing.T) {
	e := bookingGate(t, &stubBookingReadRepo{})
	token, _ := jwtutil.Issue(testSecret, time.Hour, testAdminSubject, "user")
	for _, r := range bookingRoutes() {
		req := httptest.NewRequest(r.method, r.path, nil)
		req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with role=user: %d, want 403", r.method, r.path, rec.Code)
		}
	}
}

func TestGetBookingMalformedIDIs400(t *testing.T) {
	e := bookingGate(t, &stubBookingReadRepo{})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, bookingRequest(t, http.MethodGet, "/api/v1/admin/bookings/not-a-uuid", ""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestCancelMalformedBodyIs400(t *testing.T) {
	e := bookingGate(t, &stubBookingReadRepo{})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, bookingRequest(t, http.MethodPatch, "/api/v1/admin/bookings/"+testBookingID+"/cancel", "{not json"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

// The reason is validated before the transaction opens, so this never
// reaches the (nil) pool — an empty reason is a 422, not a 500.
func TestCancelWithoutReasonIs422(t *testing.T) {
	e := bookingGate(t, &stubBookingReadRepo{})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, bookingRequest(t, http.MethodPatch, "/api/v1/admin/bookings/"+testBookingID+"/cancel", `{"cancellation_reason":"  "}`))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want 422; body %s", rec.Code, rec.Body.String())
	}
}

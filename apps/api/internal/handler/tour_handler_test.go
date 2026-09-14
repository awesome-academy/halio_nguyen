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

// Stub repos below embed the real interface (any unexercised method panics
// loudly) — these tests only ever reach the A0 gate or the handler's own
// Bind/param validation, never a repository call, except stubTourRepo.List
// for the one envelope-shape test (mirrors category_handler_test.go). Shared
// by tour_image_handler_test.go and tour_schedule_handler_test.go.

type stubTourRepo struct {
	repository.TourRepository
	items []domain.Tour
}

func (s stubTourRepo) List(context.Context, repository.DB, repository.TourListParams) ([]domain.Tour, int64, error) {
	return s.items, int64(len(s.items)), nil
}

type stubTourImageRepo struct{ repository.TourImageRepository }
type stubTourScheduleRepo struct {
	repository.TourScheduleRepository
}
type stubBookingRepo struct{ repository.BookingRepository }

const (
	testTourID  = "33333333-3333-3333-3333-333333333001"
	testImageID = "33333333-3333-3333-3333-333333333002"
	testSchedID = "33333333-3333-3333-3333-333333333003"
)

// tourGate wires all three F003 handlers (A1-A13) behind the same A0 chain
// router.New installs, so the gate is proven by request, not inspection
// (phase-04 Security Considerations).
func tourGate(t *testing.T, tourRepo stubTourRepo) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	gated := e.Group("/api/v1/admin")
	gated.Use(jwtGate(t))

	tourSvc := service.NewTourService(nil, tourRepo, stubCategoryRepo{}, stubTourImageRepo{}, stubTourScheduleRepo{}, stubBookingRepo{})
	imgSvc := service.NewTourImageService(nil, tourRepo, stubTourImageRepo{})
	schSvc := service.NewTourScheduleService(nil, tourRepo, stubTourScheduleRepo{}, stubBookingRepo{})

	NewTourHandler(tourSvc).RegisterRoutes(gated)
	NewTourImageHandler(imgSvc).RegisterRoutes(gated)
	NewTourScheduleHandler(schSvc).RegisterRoutes(gated)
	return e
}

// tourRoutes lists all 13 F003 actions (SC: every one 401 without a cookie,
// 403 with a role:"user" token).
func tourRoutes() []struct{ method, path string } {
	return []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/tours"},
		{http.MethodGet, "/api/v1/admin/tours/" + testTourID},
		{http.MethodPost, "/api/v1/admin/tours"},
		{http.MethodPut, "/api/v1/admin/tours/" + testTourID},
		{http.MethodPatch, "/api/v1/admin/tours/" + testTourID + "/status"},
		{http.MethodDelete, "/api/v1/admin/tours/" + testTourID},
		{http.MethodPost, "/api/v1/admin/tours/" + testTourID + "/images"},
		{http.MethodPut, "/api/v1/admin/tours/" + testTourID + "/images/" + testImageID},
		{http.MethodDelete, "/api/v1/admin/tours/" + testTourID + "/images/" + testImageID},
		{http.MethodPost, "/api/v1/admin/tours/" + testTourID + "/schedules"},
		{http.MethodPut, "/api/v1/admin/tours/" + testTourID + "/schedules/" + testSchedID},
		{http.MethodPatch, "/api/v1/admin/tours/" + testTourID + "/schedules/" + testSchedID + "/status"},
		{http.MethodDelete, "/api/v1/admin/tours/" + testTourID + "/schedules/" + testSchedID},
	}
}

func TestAllTourRoutesRejectMissingCookie(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	for _, r := range tourRoutes() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without cookie: %d, want 401", r.method, r.path, rec.Code)
		}
	}
}

func TestAllTourRoutesRejectNonAdminRole(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	token, _ := jwtutil.Issue(testSecret, time.Hour, "u1", "user")
	for _, r := range tourRoutes() {
		req := httptest.NewRequest(r.method, r.path, nil)
		req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with role=user: %d, want 403", r.method, r.path, rec.Code)
		}
	}
}

func TestListToursReturnsPaginatedEnvelope(t *testing.T) {
	e := tourGate(t, stubTourRepo{items: []domain.Tour{{ID: uuid.New(), Title: "Sample"}}})
	req := adminRequest(t, http.MethodGet, "/api/v1/admin/tours?page_size=5", "")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	for _, key := range []string{`"items"`, `"total":1`, `"page":1`, `"page_size":5`} {
		if !strings.Contains(rec.Body.String(), key) {
			t.Errorf("body missing %s: %s", key, rec.Body.String())
		}
	}
}

func TestGetTourMalformedIDIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodGet, "/api/v1/admin/tours/not-a-uuid", "")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestCreateTourMalformedBodyIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodPost, "/api/v1/admin/tours", "{not json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestUpdateTourStatusRequiresStatus(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodPatch, "/api/v1/admin/tours/"+testTourID+"/status", "{}")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

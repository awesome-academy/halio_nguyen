package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{JWT: config.JWTConfig{Secret: "test-secret-at-least-32-bytes-long!"}}
}

// newTestEcho wires apperror.Handler the same way main.go does — router.New
// itself only builds routes/groups, so a test that skips this step would
// see echo's generic default error body instead of the real API contract.
func newTestEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	return e
}

// TestUnbuiltAdminRouteRequiresAuth proves A0 runs ahead of routing-into-
// handler logic for the whole /api/v1/admin group (phase-02 Success
// Criteria) — a path no phase has built yet (categories ships in Phase 3)
// must still 401 unauthenticated, never 404, so a later phase cannot
// accidentally ship a route outside the gate without a test catching it.
func TestUnbuiltAdminRouteRequiresAuth(t *testing.T) {
	e := newTestEcho()
	New(e, testConfig(), Deps{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/categories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d (401, not 404)", rec.Code, http.StatusUnauthorized)
	}
}

func TestInfoRouteStaysPublic(t *testing.T) {
	e := newTestEcho()
	New(e, testConfig(), Deps{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// stubCategoryRepo overrides only List; any other call panics loudly.
type stubCategoryRepo struct {
	repository.CategoryRepository
	items []repository.CategoryListItem
}

func (s stubCategoryRepo) List(context.Context, repository.DB, repository.CategoryListParams) ([]repository.CategoryListItem, int64, error) {
	return s.items, int64(len(s.items)), nil
}

// categoryGate mirrors router.New's A0 chain: the handler test proves the
// gate by request, not by inspection (phase-03 Security Considerations).
func categoryGate(t *testing.T, repo repository.CategoryRepository) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	gated := e.Group("/api/v1/admin")
	gated.Use(jwtGate(t))
	NewCategoryHandler(service.NewCategoryService(nil, repo)).RegisterRoutes(gated)
	return e
}

func categoryRoutes() []struct{ method, path string } {
	return []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/categories"},
		{http.MethodPost, "/api/v1/admin/categories"},
		{http.MethodPut, "/api/v1/admin/categories/22222222-2222-2222-2222-222222222001"},
		{http.MethodPatch, "/api/v1/admin/categories/22222222-2222-2222-2222-222222222001/sort-order"},
		{http.MethodDelete, "/api/v1/admin/categories/22222222-2222-2222-2222-222222222001"},
	}
}

func TestAllCategoryRoutesRejectMissingCookie(t *testing.T) {
	e := categoryGate(t, stubCategoryRepo{})
	for _, r := range categoryRoutes() {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without cookie: %d, want 401", r.method, r.path, rec.Code)
		}
	}
}

func TestAllCategoryRoutesRejectNonAdminRole(t *testing.T) {
	e := categoryGate(t, stubCategoryRepo{})
	token, _ := jwtutil.Issue(testSecret, time.Hour, "u1", "user")
	for _, r := range categoryRoutes() {
		req := httptest.NewRequest(r.method, r.path, nil)
		req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with role=user: %d, want 403", r.method, r.path, rec.Code)
		}
	}
}

func TestListCategoriesReturnsPaginatedEnvelope(t *testing.T) {
	e := categoryGate(t, stubCategoryRepo{items: []repository.CategoryListItem{{TourCount: 2}}})
	req := adminRequest(t, http.MethodGet, "/api/v1/admin/categories?sort_by=name;DROP%20TABLE&page_size=5", "")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	for _, key := range []string{`"items"`, `"total":1`, `"page":1`, `"page_size":5`, `"tour_count":2`} {
		if !strings.Contains(rec.Body.String(), key) {
			t.Errorf("body missing %s: %s", key, rec.Body.String())
		}
	}
}

func TestCreateCategoryMalformedBodyIs400(t *testing.T) {
	e := categoryGate(t, stubCategoryRepo{})
	req := adminRequest(t, http.MethodPost, "/api/v1/admin/categories", "{not json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestReorderRequiresSortOrder(t *testing.T) {
	e := categoryGate(t, stubCategoryRepo{})
	req := adminRequest(t, http.MethodPatch, "/api/v1/admin/categories/22222222-2222-2222-2222-222222222001/sort-order", "{}")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func adminRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	token, err := jwtutil.Issue(testSecret, time.Hour, "u1", "admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
	return req
}

// jwtGate is the same echojwt → RequireAdminRole pair router.New installs
// (the handler package cannot import router without a cycle).
func jwtGate(t *testing.T) echo.MiddlewareFunc {
	t.Helper()
	cfg := &config.Config{JWT: config.JWTConfig{Secret: testSecret}}
	jwtMw := echojwt.WithConfig(adminmw.NewJWTConfig(cfg))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return jwtMw(adminmw.RequireAdminRole(next))
	}
}

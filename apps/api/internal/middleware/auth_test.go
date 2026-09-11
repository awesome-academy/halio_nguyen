package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
)

const testSecret = "test-secret-at-least-32-bytes-long!"

func testCfg() *config.Config {
	return &config.Config{JWT: config.JWTConfig{Secret: testSecret}}
}

// gateChain assembles the A0 chain (echojwt then RequireAdminRole) exactly
// as router.go will, ending in a handler that proves it ran.
func gateChain() echo.HandlerFunc {
	handler := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}
	jwtMw := echojwt.WithConfig(NewJWTConfig(testCfg()))
	return jwtMw(RequireAdminRole(handler))
}

func newContext(cookieValue string) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: cookieValue})
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func runGate(t *testing.T, cookieValue string) error {
	t.Helper()
	return gateChain()(newContext(cookieValue))
}

func TestGateRejectsNoCookie(t *testing.T) {
	err := runGate(t, "")
	assertUnauthorized(t, err)
}

func TestGateRejectsBadSignature(t *testing.T) {
	token, err := jwtutil.Issue("a-totally-different-secret-32-b!", time.Hour, "u1", "admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	assertUnauthorized(t, runGate(t, token))
}

func TestGateRejectsExpiredToken(t *testing.T) {
	token, err := jwtutil.Issue(testSecret, -time.Minute, "u1", "admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	assertUnauthorized(t, runGate(t, token))
}

func TestGateRejectsNonAdminRole(t *testing.T) {
	token, err := jwtutil.Issue(testSecret, time.Hour, "u1", "user")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	err = runGate(t, token)
	appErr, ok := err.(*apperror.Error)
	if !ok {
		t.Fatalf("expected *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.Code != apperror.CodeForbidden {
		t.Errorf("expected forbidden, got %v", appErr.Code)
	}
}

func TestGateAllowsValidAdminToken(t *testing.T) {
	token, err := jwtutil.Issue(testSecret, time.Hour, "u1", "admin")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if err := runGate(t, token); err != nil {
		t.Fatalf("expected the gate to pass a valid admin token through, got error: %v", err)
	}
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	appErr, ok := err.(*apperror.Error)
	if !ok {
		t.Fatalf("expected *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("expected unauthorized, got %v", appErr.Code)
	}
}

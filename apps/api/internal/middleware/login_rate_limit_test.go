package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
)

// hitLogin runs one request through the limiter and returns the written
// status. Echo's RateLimiter reports a deny via c.Error (which invokes the
// HTTPErrorHandler) and returns nil, so the recorder — not the returned
// error — is where the 429 shows up.
func hitLogin(t *testing.T, limiter echo.MiddlewareFunc, ip string) int {
	t.Helper()
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = ip + ":12345"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := limiter(func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	if err := handler(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
	return rec.Code
}

// TestLoginRateLimiterBlocksSixthAttemptPerIP pins D3's per-IP limb: with
// 5/min + burst 5, five rapid attempts pass and the sixth is a 429, while
// a different client IP is unaffected.
func TestLoginRateLimiterBlocksSixthAttemptPerIP(t *testing.T) {
	limiter := NewLoginRateLimiter(&config.SecurityConfig{LoginRateLimitPerMinute: 5, LoginRateLimitBurst: 5})

	for i := 1; i <= 5; i++ {
		if code := hitLogin(t, limiter, "203.0.113.1"); code != http.StatusOK {
			t.Fatalf("attempt %d: status %d, want 200", i, code)
		}
	}
	if code := hitLogin(t, limiter, "203.0.113.1"); code != http.StatusTooManyRequests {
		t.Fatalf("6th attempt: status %d, want 429", code)
	}
	if code := hitLogin(t, limiter, "203.0.113.2"); code != http.StatusOK {
		t.Fatalf("different IP: status %d, want 200 (must not be throttled)", code)
	}
}

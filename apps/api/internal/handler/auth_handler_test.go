package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

const testSecret = "test-secret-at-least-32-bytes-long!"

// adminHash is a real bcrypt cost-10 hash of "Correct@123456".
const adminHash = "$2a$10$..QrZV9la86HLhpx6rA/W.qDWZXcK2mB/CyFMa0SrejjnSgZO4DAu"

type stubUserRepo struct {
	user *domain.User
	err  error
}

func (s stubUserRepo) FindActiveAdminByEmail(ctx context.Context, db repository.DB, email string) (*domain.User, error) {
	return s.user, s.err
}

func (s stubUserRepo) FindByID(ctx context.Context, db repository.DB, id uuid.UUID) (*domain.User, error) {
	return s.user, s.err
}

type stubActivityRepo struct{}

func (stubActivityRepo) Insert(ctx context.Context, db repository.DB, log repository.NewActivityLog) error {
	return nil
}

func testConfig() *config.Config {
	return &config.Config{
		JWT:      config.JWTConfig{Secret: testSecret, AccessExpiryHours: time.Hour},
		Security: config.SecurityConfig{CookieSecure: false},
	}
}

func newHandler(user *domain.User, lookupErr error) *AuthHandler {
	throttle := service.NewEmailThrottle(1000, time.Minute)
	svc := service.NewAuthService(nil, stubUserRepo{user: user, err: lookupErr}, stubActivityRepo{}, throttle, testSecret, time.Hour)
	return NewAuthHandler(svc, testConfig())
}

func adminUser() *domain.User {
	hash := adminHash
	return &domain.User{ID: uuid.New(), Email: "admin@sunbooking.com", PasswordHash: &hash, FullName: "System Administrator", Role: domain.RoleAdmin, IsActive: true}
}

func TestLoginSetsSessionCookieAttributes(t *testing.T) {
	e := echo.New()
	h := newHandler(adminUser(), nil)

	body, _ := json.Marshal(map[string]string{"email": "admin@sunbooking.com", "password": "Correct@123456"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	setCookie := rec.Header().Get("Set-Cookie")
	for _, want := range []string{"sun_admin_token=", "Path=/", "HttpOnly", "SameSite=Lax", "Max-Age=3600"} {
		if !strings.Contains(setCookie, want) {
			t.Errorf("Set-Cookie %q missing %q", setCookie, want)
		}
	}

	var resp loginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp.RedirectTo != "/admin/dashboard" {
		t.Errorf("RedirectTo = %q, want default dashboard", resp.RedirectTo)
	}
	if resp.User.Email != "admin@sunbooking.com" {
		t.Errorf("User.Email = %q, want admin@sunbooking.com", resp.User.Email)
	}
}

func TestLoginRejectsMissingFields(t *testing.T) {
	e := echo.New()
	h := newHandler(adminUser(), nil)

	body, _ := json.Marshal(map[string]string{"email": "", "password": ""})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c := e.NewContext(req, httptest.NewRecorder())

	err := h.Login(c)
	if err == nil {
		t.Fatal("expected an error for empty email/password")
	}
}

func TestLogoutClearsCookieWithNoPriorSession(t *testing.T) {
	e := echo.New()
	h := newHandler(nil, repository.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Logout(c); err != nil {
		t.Fatalf("Logout returned error with no prior session: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	setCookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "Max-Age=0") {
		t.Errorf("expected logout to clear the cookie (Max-Age=0), got %q", setCookie)
	}
}

func TestMeReturnsProfileWithNoPasswordHash(t *testing.T) {
	e := echo.New()
	user := adminUser()
	h := newHandler(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setGateClaims(c, user.ID.String(), domain.RoleAdmin)

	if err := h.Me(c); err != nil {
		t.Fatalf("Me returned error: %v", err)
	}
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatalf("response leaked password_hash: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), user.Email) {
		t.Errorf("expected response to contain the admin's email, got %s", rec.Body.String())
	}
}

func TestMeRejectsWhenGateDidNotRun(t *testing.T) {
	e := echo.New()
	h := newHandler(adminUser(), nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	if err := h.Me(c); err == nil {
		t.Fatal("expected Me to reject a request with no gate-verified claims")
	}
}

// setGateClaims stores a *jwt.Token the way echojwt does, so
// adminmw.ClaimsFrom(c) resolves exactly as it would after the real A0
// gate ran.
func setGateClaims(c echo.Context, userID, role string) {
	claims := &jwtutil.AdminClaims{UserID: userID, Role: role}
	token := &jwt.Token{Claims: claims, Valid: true}
	c.Set("user", token)
}

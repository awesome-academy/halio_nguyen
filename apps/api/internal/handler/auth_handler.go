// Package handler holds the HTTP entry points for the admin portal. Each
// handler only binds/validates request shape, calls its service, and maps
// the result onto the wire — business rules stay in internal/service.
package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// sessionCookieName is the HttpOnly cookie carrying the admin JWT
// (decisions.md A1). middleware.NewJWTConfig reads the same name.
const sessionCookieName = "sun_admin_token"

// AuthHandler is the HTTP entry point for F001: login, logout, me.
type AuthHandler struct {
	service *service.AuthService
	cfg     *config.Config
}

// NewAuthHandler builds an AuthHandler bound to svc for the business logic
// and cfg for the cookie attributes (Secure flag, access-token TTL).
func NewAuthHandler(svc *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{service: svc, cfg: cfg}
}

// RegisterRoutes wires login/logout onto ungated (they run before any
// session exists — see phase-02 Key Insights on why logout must sit
// outside the A0 gate) and me onto gated. loginRateLimiter is the per-IP
// throttle (D3) — attached to login only, never logout/me.
func (h *AuthHandler) RegisterRoutes(gated, ungated *echo.Group, loginRateLimiter echo.MiddlewareFunc) {
	ungated.POST("/auth/login", h.Login, loginRateLimiter)
	ungated.POST("/auth/logout", h.Logout)
	gated.GET("/auth/me", h.Me)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// From is the same-origin admin path middleware.ts recorded before
	// redirecting to /admin/login (DEC-001); AuthService.ValidateRedirect
	// is the actual gate, so an attacker-controlled value here can only
	// ever resolve to itself (if valid) or the dashboard fallback.
	From string `json:"from"`
}

type adminProfile struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

func toAdminProfile(u *domain.User) adminProfile {
	return adminProfile{ID: u.ID.String(), Email: u.Email, FullName: u.FullName, Role: u.Role}
}

type loginResponse struct {
	User       adminProfile `json:"user"`
	RedirectTo string       `json:"redirect_to"`
}

// Login validates shape only — AuthService.Login owns every credential and
// anti-enumeration rule (BR-001).
func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		return apperror.NewBadRequest("Email and password are required")
	}

	result, err := h.service.Login(c.Request().Context(), service.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		From:      req.From,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
	if err != nil {
		return err
	}

	h.setSessionCookie(c, result.Token, int(h.cfg.JWT.AccessExpiryHours.Seconds()))

	return c.JSON(http.StatusOK, loginResponse{
		User:       toAdminProfile(result.User),
		RedirectTo: result.RedirectTo,
	})
}

// Logout is registered OUTSIDE the A0 gate (see RegisterRoutes) so it stays
// idempotent with no session cookie present (SC-003). It clears the cookie
// unconditionally and only attempts the best-effort activity_logs write
// when a token happens to parse — a missing/invalid/expired cookie is not
// an error here, since there is nothing to log or reject.
func (h *AuthHandler) Logout(c echo.Context) error {
	h.setSessionCookie(c, "", -1)

	if userID := h.userIDFromCookie(c); userID != nil {
		h.service.Logout(c.Request().Context(), userID, c.RealIP(), c.Request().UserAgent())
	}

	return c.NoContent(http.StatusOK)
}

// Me returns the profile for the identity the A0 gate already verified —
// this handler trusts adminmw.ClaimsFrom rather than re-parsing the cookie.
func (h *AuthHandler) Me(c echo.Context) error {
	claims, ok := adminmw.ClaimsFrom(c)
	if !ok {
		return apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}

	user, err := h.service.Me(c.Request().Context(), claims.UserID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toAdminProfile(user))
}

// setSessionCookie is the one place cookie attributes are written so login
// and logout cannot drift on them (phase-02 step 9). maxAge follows
// net/http.Cookie semantics: >0 sets an expiry, <0 deletes the cookie now.
func (h *AuthHandler) setSessionCookie(c echo.Context, token string, maxAge int) {
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// userIDFromCookie best-effort parses the session cookie for its subject.
// Any failure (no cookie, expired, bad signature) yields nil, not an
// error — logout has nothing to reject.
func (h *AuthHandler) userIDFromCookie(c echo.Context) *uuid.UUID {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	claims, err := jwtutil.Parse(h.cfg.JWT.Secret, cookie.Value)
	if err != nil {
		return nil
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil
	}
	return &id
}

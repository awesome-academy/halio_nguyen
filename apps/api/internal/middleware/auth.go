// Package middleware holds the admin portal's cross-cutting HTTP guards: the
// A0 JWT + RBAC gate every /api/v1/admin/* route sits behind, and the login
// rate limiter. Business rules stay in internal/service.
package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
)

// unauthorizedMessage is the single message every A0 rejection returns,
// regardless of cause (missing cookie, bad signature, expired token) — the
// caller learns only that the session is not valid, never why.
const unauthorizedMessage = "Your session has expired or is invalid. Please log in again."

// NewJWTConfig builds the echojwt.Config for the admin session gate: it
// reads the token from the sun_admin_token cookie, decodes it into
// *jwtutil.AdminClaims, and pins SigningMethod to HS256 so a forged
// alg:none or RS256-confusion token cannot pass verification.
func NewJWTConfig(cfg *config.Config) echojwt.Config {
	return echojwt.Config{
		TokenLookup:   "cookie:sun_admin_token",
		SigningKey:    []byte(cfg.JWT.Secret),
		SigningMethod: echojwt.AlgorithmHS256,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &jwtutil.AdminClaims{}
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return apperror.NewUnauthorized(unauthorizedMessage)
		},
	}
}

// RequireAdminRole runs after the echojwt gate and rejects any valid token
// whose role claim is not "admin" (FR-601). It type-asserts defensively —
// `, ok` rather than a bare assertion — so an echo-jwt/golang-jwt version
// skew that changes the stored context type fails as a 401, never a panic.
func RequireAdminRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := ClaimsFrom(c)
		if !ok {
			return apperror.NewUnauthorized(unauthorizedMessage)
		}
		if claims.Role != "admin" {
			return apperror.NewForbidden("This account does not have admin access.")
		}
		return next(c)
	}
}

// ClaimsFrom extracts the verified *jwtutil.AdminClaims the A0 gate stored
// on the request context. ok is false whenever the gate did not run, the
// stored value is not a *jwt.Token (version skew — see R6), or its Claims
// are not *jwtutil.AdminClaims.
func ClaimsFrom(c echo.Context) (*jwtutil.AdminClaims, bool) {
	raw, ok := c.Get("user").(*jwt.Token)
	if !ok || raw == nil {
		return nil, false
	}
	claims, ok := raw.Claims.(*jwtutil.AdminClaims)
	if !ok {
		return nil, false
	}
	return claims, true
}

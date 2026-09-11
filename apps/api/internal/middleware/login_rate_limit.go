package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"golang.org/x/time/rate"
)

// rateLimiterExpiry is how long an idle per-IP bucket is kept before the
// store's janitor reclaims it.
const rateLimiterExpiry = 3 * time.Minute

// NewLoginRateLimiter builds the per-IP limb of D3's login throttle (the
// per-email limb lives in service.EmailThrottle, since the body isn't
// readable yet when this middleware runs). It is attached to the login
// route only — see router.go.
func NewLoginRateLimiter(cfg *config.SecurityConfig) echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(cfg.LoginRateLimitPerMinute) / 60,
			Burst:     cfg.LoginRateLimitBurst,
			ExpiresIn: rateLimiterExpiry,
		}),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return apperror.NewInternal(err)
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return apperror.NewRateLimited("Too many login attempts. Please try again later.")
		},
	})
}

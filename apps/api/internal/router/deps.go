package router

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/handler"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// emailThrottleWindow is D3's fixed window for the per-email login limb.
const emailThrottleWindow = time.Minute

// Build constructs every repository → service → handler chain the router
// registers. It lives here (not in main.go) so main stays frozen: each
// later phase adds its own three lines to this function and one field to
// Deps.
func Build(cfg *config.Config, pool *pgxpool.Pool) Deps {
	userRepo := repository.NewUserRepository()
	activityRepo := repository.NewActivityLogRepository()
	emailThrottle := service.NewEmailThrottle(cfg.Security.LoginRateLimitPerMinute, emailThrottleWindow)
	authService := service.NewAuthService(pool, userRepo, activityRepo, emailThrottle, cfg.JWT.Secret, cfg.JWT.AccessExpiryHours)

	categoryService := service.NewCategoryService(pool, repository.NewCategoryRepository())

	return Deps{
		AuthHandler: handler.NewAuthHandler(authService, cfg),
		Categories:  handler.NewCategoryHandler(categoryService),
	}
}

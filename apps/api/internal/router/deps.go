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

	categoryRepo := repository.NewCategoryRepository()
	categoryService := service.NewCategoryService(pool, categoryRepo)

	tourRepo := repository.NewTourRepository()
	tourImageRepo := repository.NewTourImageRepository()
	tourScheduleRepo := repository.NewTourScheduleRepository()
	bookingRepo := repository.NewBookingRepository()

	tourService := service.NewTourService(pool, tourRepo, categoryRepo, tourImageRepo, tourScheduleRepo, bookingRepo)
	tourImageService := service.NewTourImageService(pool, tourRepo, tourImageRepo)
	tourScheduleService := service.NewTourScheduleService(pool, tourRepo, tourScheduleRepo, bookingRepo)
	bookingService := service.NewBookingService(pool, bookingRepo, tourScheduleRepo, activityRepo)

	userAdminRepo := repository.NewUserAdminRepository()
	userAdminService := service.NewUserAdminService(pool, userAdminRepo, bookingRepo)

	reviewAdminRepo := repository.NewReviewAdminRepository()
	commentAdminRepo := repository.NewCommentAdminRepository()
	reviewModerationService := service.NewReviewModerationService(pool, reviewAdminRepo)
	commentModerationService := service.NewCommentModerationService(pool, commentAdminRepo)

	revenueRepo := repository.NewRevenueRepository()
	// The runner needs the raw pool: it calls Acquire to own one session for
	// the whole lock/refresh/unlock cycle, which repository.DB deliberately
	// does not expose.
	revenueRunner := service.NewRevenueRefreshRunner(service.PoolConnSource{Pool: pool})
	revenueService := service.NewRevenueService(pool, revenueRepo, revenueRunner)

	return Deps{
		AuthHandler:   handler.NewAuthHandler(authService, cfg),
		Categories:    handler.NewCategoryHandler(categoryService),
		Tours:         handler.NewTourHandler(tourService),
		TourImages:    handler.NewTourImageHandler(tourImageService),
		TourSchedules: handler.NewTourScheduleHandler(tourScheduleService),
		Bookings:      handler.NewBookingHandler(bookingService),
		Users:         handler.NewUserAdminHandler(userAdminService),
		Reviews:       handler.NewReviewAdminHandler(reviewModerationService, commentModerationService),
		Revenue:       handler.NewRevenueHandler(revenueService),
	}
}

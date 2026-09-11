// Package router owns the route registry for the API. main.go calls New
// once and is frozen after that — every later phase registers its own
// routes here instead of editing main.go again.
package router

import (
	"net/http"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/handler"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
)

// Deps carries the shared dependencies route groups need. Each later phase
// adds the fields its own handlers require.
type Deps struct {
	AuthHandler *handler.AuthHandler
	Categories  *handler.CategoryHandler
}

// New wires the API's /api/v1 route group onto e. It preserves the existing
// /api/v1/info route and builds the /api/v1/admin group's A0 gate — echojwt
// (cookie sun_admin_token) then RequireAdminRole — every admin-portal
// feature registers against. login/logout are registered on the sibling
// ungated group instead (they run before any session exists).
func New(e *echo.Echo, cfg *config.Config, deps Deps) {
	v1 := e.Group("/api/v1")
	v1.GET("/info", infoHandler)

	ungatedAdmin := v1.Group("/admin")
	gatedAdmin := v1.Group("/admin")
	gatedAdmin.Use(echojwt.WithConfig(adminmw.NewJWTConfig(cfg)))
	gatedAdmin.Use(adminmw.RequireAdminRole)

	// A wildcard fallback so the gate runs even for a not-yet-built admin
	// route (FR-601's "before any handler or resource lookup" holds for the
	// whole group, not just registered handlers). A later phase's more
	// specific route on gatedAdmin always wins over this catch-all.
	gatedAdmin.Any("/*", notFoundHandler)

	if deps.AuthHandler != nil {
		loginRateLimiter := adminmw.NewLoginRateLimiter(&cfg.Security)
		deps.AuthHandler.RegisterRoutes(gatedAdmin, ungatedAdmin, loginRateLimiter)
	}
	if deps.Categories != nil {
		deps.Categories.RegisterRoutes(gatedAdmin)
	}
}

func infoHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"version": "1.0.0",
		"service": "SUN Booking Tours API",
	})
}

func notFoundHandler(c echo.Context) error {
	return apperror.NewNotFound("Not found")
}

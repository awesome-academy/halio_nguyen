// Package router owns the route registry for the API. main.go calls New
// once and is frozen after that — every later phase registers its own
// routes here instead of editing main.go again.
package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
)

// Deps carries the shared dependencies route groups need (repositories,
// services, middleware). It is empty in this phase; later phases add fields
// as their handlers require them.
type Deps struct{}

// New wires the API's /api/v1 route group onto e. It preserves the existing
// /api/v1/info route and creates the /api/v1/admin group every admin-portal
// feature registers against. The admin group's middleware chain (JWT auth +
// RBAC) is empty here and filled in by Phase 2.
func New(e *echo.Echo, cfg *config.Config, deps Deps) {
	v1 := e.Group("/api/v1")
	v1.GET("/info", infoHandler)

	// admin is the group every admin-portal route sits behind. Later phases
	// add their middleware chain and call xxxHandler.RegisterRoutes(admin).
	v1.Group("/admin")
}

func infoHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"version": "1.0.0",
		"service": "SUN Booking Tours API",
	})
}

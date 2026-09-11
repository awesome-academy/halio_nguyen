package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/config"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/db"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/router"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err, "dsn_host", cfg.Database.Host)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("connected to PostgreSQL successfully", "db", cfg.Database.DBName)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = apperror.Handler

	// Trust X-Forwarded-For only from the explicit TRUSTED_PROXY_CIDRS hops
	// (the Next.js proxy) so the per-IP login throttle sees the real client
	// and no other internal host can spoof it.
	ipExtractor, err := adminmw.NewIPExtractor(cfg.Security.TrustedProxyCIDRs)
	if err != nil {
		logger.Error("invalid trusted proxy configuration", "error", err)
		os.Exit(1)
	}
	e.IPExtractor = ipExtractor

	// Middlewares
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// localhost:3000 is a dev-only convenience for running the Next.js app
	// without the rewrites proxy in front of it; a production CORS list must
	// not carry it (red-team finding: an attacker-hosted localhost page would
	// otherwise be CORS-permitted against the live API in every environment).
	allowOrigins := []string{cfg.App.AllowedOrigins}
	if cfg.App.Env == "development" {
		allowOrigins = append(allowOrigins, "http://localhost:3000")
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		pingCtx, pingCancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer pingCancel()

		if err := pool.Ping(pingCtx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "unhealthy",
				"db":     "disconnected",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"app":       cfg.App.Name,
			"env":       cfg.App.Env,
			"timestamp": time.Now().UTC(),
		})
	})

	// Route registry: main.go is frozen after this — every later phase wires
	// its chain in router.Build and its routes in router.New, not here.
	router.New(e, cfg, router.Build(cfg, pool))

	// Graceful shutdown server
	go func() {
		addr := ":" + cfg.App.Port
		logger.Info("starting server", "addr", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server startup failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server exited cleanly")
}

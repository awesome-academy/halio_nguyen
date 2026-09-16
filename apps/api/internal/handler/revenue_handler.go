package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// RevenueHandler is the HTTP entry point for F007 (A1–A3).
type RevenueHandler struct {
	svc *service.RevenueService
}

// NewRevenueHandler builds the handler over its service.
func NewRevenueHandler(svc *service.RevenueService) *RevenueHandler {
	return &RevenueHandler{svc: svc}
}

// RegisterRoutes attaches all three routes to the A0-gated admin group.
// Revenue is commercially sensitive and there is no role narrower than
// admin in the schema, so every route sits behind the gate (BR-003).
func (h *RevenueHandler) RegisterRoutes(gated *echo.Group) {
	gated.GET("/revenue/daily", h.ListDaily)
	gated.GET("/revenue/monthly", h.ListMonthly)
	gated.POST("/revenue/refresh", h.Refresh)
}

// revenueListResponse is the one envelope both reads return. It is the
// shared Paginated shape plus last_refreshed_at, which is null when this
// process has not completed a refresh (DEC-001 / L3) — the UI renders that
// null as "unknown" rather than inventing a time.
type revenueListResponse[T any] struct {
	Items           []T        `json:"items"`
	Total           int64      `json:"total"`
	Page            int        `json:"page"`
	PageSize        int        `json:"page_size"`
	LastRefreshedAt *time.Time `json:"last_refreshed_at"`
}

func (h *RevenueHandler) ListDaily(c echo.Context) error {
	var p repository.DailyRevenueParams
	p.Page, _ = strconv.Atoi(c.QueryParam("page"))
	p.PageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	p.SortBy = c.QueryParam("sort_by")
	p.SortDir = c.QueryParam("sort_dir")

	page, err := h.svc.ListDaily(c.Request().Context(), p, c.QueryParam("from"), c.QueryParam("to"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, revenueListResponse[any]{
		Items:           toAnySlice(page.Items),
		Total:           page.Total,
		Page:            page.Page,
		PageSize:        page.PageSize,
		LastRefreshedAt: page.LastRefreshedAt,
	})
}

// ListMonthly returns the whole range in one response: the monthly view
// groups by month + category only, so the row count is bounded by
// months x categories and pagination would buy nothing. The envelope stays
// the shared one so every admin list looks the same on the wire.
func (h *RevenueHandler) ListMonthly(c echo.Context) error {
	page, err := h.svc.ListMonthly(
		c.Request().Context(),
		c.QueryParam("sort_by"), c.QueryParam("sort_dir"),
		c.QueryParam("from"), c.QueryParam("to"),
	)
	if err != nil {
		return err
	}
	total := int64(len(page.Items))
	return c.JSON(http.StatusOK, revenueListResponse[any]{
		Items:           toAnySlice(page.Items),
		Total:           total,
		Page:            repository.DefaultPage,
		PageSize:        len(page.Items),
		LastRefreshedAt: page.LastRefreshedAt,
	})
}

// Refresh is A3. It answers before the refresh finishes — always 202 on a
// started run, 409 when one is already in flight (BR-001: rejected, never
// queued).
func (h *RevenueHandler) Refresh(c echo.Context) error {
	claims, ok := adminmw.ClaimsFrom(c)
	if !ok {
		return apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}

	started, err := h.svc.TriggerRefresh(c.Request().Context(), claims.UserID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if !started {
		return c.JSON(http.StatusConflict, map[string]string{"status": "already_in_progress"})
	}
	return c.JSON(http.StatusAccepted, map[string]string{"status": "refresh_started"})
}

// toAnySlice keeps one response struct for both row types. A nil slice must
// serialize as [] and not null, so the client never has to guard for it.
func toAnySlice[T any](items []T) []any {
	out := make([]any, 0, len(items))
	for _, it := range items {
		out = append(out, it)
	}
	return out
}

package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// BookingHandler is the HTTP entry point for F004's two reads and three
// guarded transitions (A1-A5). There is no create, update or delete here by
// design: the customer site owns booking creation, and nothing in this
// feature writes the payments table.
type BookingHandler struct {
	svc *service.BookingService
}

// NewBookingHandler builds the handler over its service.
func NewBookingHandler(svc *service.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// RegisterRoutes attaches A1-A5 to the A0-gated admin group.
func (h *BookingHandler) RegisterRoutes(gated *echo.Group) {
	gated.GET("/bookings", h.List)
	gated.GET("/bookings/:id", h.Get)
	gated.PATCH("/bookings/:id/confirm", h.Confirm)
	gated.PATCH("/bookings/:id/cancel", h.Cancel)
	gated.PATCH("/bookings/:id/complete", h.Complete)
}

// bookingCancelRequest is A4's body. The reason is required; the service
// trims and length-caps it (BR-004).
type bookingCancelRequest struct {
	CancellationReason string `json:"cancellation_reason"`
}

func (h *BookingHandler) List(c echo.Context) error {
	p, err := parseBookingListParams(c)
	if err != nil {
		return err
	}
	page, err := h.svc.List(c.Request().Context(), p)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, page)
}

func (h *BookingHandler) Get(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	booking, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, booking)
}

func (h *BookingHandler) Confirm(c echo.Context) error {
	id, actor, err := bookingActionContext(c)
	if err != nil {
		return err
	}
	updated, err := h.svc.Confirm(c.Request().Context(), id, actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *BookingHandler) Complete(c echo.Context) error {
	id, actor, err := bookingActionContext(c)
	if err != nil {
		return err
	}
	updated, err := h.svc.Complete(c.Request().Context(), id, actor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *BookingHandler) Cancel(c echo.Context) error {
	id, actor, err := bookingActionContext(c)
	if err != nil {
		return err
	}
	var req bookingCancelRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}

	cancelled, err := h.svc.Cancel(c.Request().Context(), id, service.CancelInput{
		Reason:    req.CancellationReason,
		Actor:     actor,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, cancelled)
}

// bookingActionContext resolves the two inputs every transition needs: the
// booking from the path, and the acting admin from the verified JWT subject
// (activity_logs.user_id is the actor, never the booking's customer — L1).
// A token that passed A0 but carries an unparseable subject is treated as no
// session at all rather than being attributed to a nil admin.
func bookingActionContext(c echo.Context) (bookingID, actor uuid.UUID, err error) {
	bookingID, err = pathUUID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	claims, ok := adminmw.ClaimsFrom(c)
	if !ok {
		return uuid.Nil, uuid.Nil, apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}
	actor, err = uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}
	return bookingID, actor, nil
}

package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// scheduleDateLayout is the wire format for departure_date/return_date
// (date-only, no time component — tour_schedules columns are DATE).
const scheduleDateLayout = "2006-01-02"

// TourScheduleHandler is the HTTP entry point for F003's schedule actions
// (A10-A13).
type TourScheduleHandler struct {
	svc *service.TourScheduleService
}

// NewTourScheduleHandler builds the handler over its service.
func NewTourScheduleHandler(svc *service.TourScheduleService) *TourScheduleHandler {
	return &TourScheduleHandler{svc: svc}
}

// RegisterRoutes attaches A10-A13 to the A0-gated admin group.
func (h *TourScheduleHandler) RegisterRoutes(gated *echo.Group) {
	gated.POST("/tours/:id/schedules", h.Create)
	gated.PUT("/tours/:id/schedules/:scheduleId", h.Update)
	gated.PATCH("/tours/:id/schedules/:scheduleId/status", h.UpdateStatus)
	gated.DELETE("/tours/:id/schedules/:scheduleId", h.Delete)
}

// tourScheduleRequest is A10/A11's DTO; also reused by TourHandler for A3's
// nested schedules[] array (same shape, same package). Dates are plain
// YYYY-MM-DD strings (tour_schedules columns are DATE, no time component).
type tourScheduleRequest struct {
	DepartureDate  string   `json:"departure_date"`
	ReturnDate     string   `json:"return_date"`
	AvailableSlots int      `json:"available_slots"`
	PriceOverride  *float64 `json:"price_override"`
}

func (r tourScheduleRequest) toInput() (service.TourScheduleInput, error) {
	dep, err := time.Parse(scheduleDateLayout, r.DepartureDate)
	if err != nil {
		return service.TourScheduleInput{}, apperror.NewBadRequest("departure_date must be YYYY-MM-DD")
	}
	ret, err := time.Parse(scheduleDateLayout, r.ReturnDate)
	if err != nil {
		return service.TourScheduleInput{}, apperror.NewBadRequest("return_date must be YYYY-MM-DD")
	}
	return service.TourScheduleInput{
		DepartureDate: dep, ReturnDate: ret, AvailableSlots: r.AvailableSlots, PriceOverride: r.PriceOverride,
	}, nil
}

type tourScheduleStatusRequest struct {
	Status string `json:"status"`
}

func (h *TourScheduleHandler) Create(c echo.Context) error {
	tourID, err := pathUUID(c)
	if err != nil {
		return err
	}
	var req tourScheduleRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	in, err := req.toInput()
	if err != nil {
		return err
	}
	created, err := h.svc.Create(c.Request().Context(), tourID, in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *TourScheduleHandler) Update(c echo.Context) error {
	tourID, err := pathUUID(c)
	if err != nil {
		return err
	}
	scheduleID, err := pathParamUUID(c, "scheduleId")
	if err != nil {
		return err
	}
	var req tourScheduleRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	in, err := req.toInput()
	if err != nil {
		return err
	}
	updated, err := h.svc.Update(c.Request().Context(), tourID, scheduleID, in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *TourScheduleHandler) UpdateStatus(c echo.Context) error {
	tourID, err := pathUUID(c)
	if err != nil {
		return err
	}
	scheduleID, err := pathParamUUID(c, "scheduleId")
	if err != nil {
		return err
	}
	var req tourScheduleStatusRequest
	if err := c.Bind(&req); err != nil || req.Status == "" {
		return apperror.NewBadRequest("status is required")
	}
	updated, err := h.svc.UpdateStatus(c.Request().Context(), tourID, scheduleID, req.Status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *TourScheduleHandler) Delete(c echo.Context) error {
	tourID, err := pathUUID(c)
	if err != nil {
		return err
	}
	scheduleID, err := pathParamUUID(c, "scheduleId")
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), tourID, scheduleID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

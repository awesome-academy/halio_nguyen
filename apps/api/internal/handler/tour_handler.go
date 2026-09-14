package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// TourHandler is the HTTP entry point for F003's tour CRUD + status
// transition (A1-A6).
type TourHandler struct {
	svc *service.TourService
}

// NewTourHandler builds the handler over its service.
func NewTourHandler(svc *service.TourService) *TourHandler {
	return &TourHandler{svc: svc}
}

// RegisterRoutes attaches A1-A6 to the A0-gated admin group.
func (h *TourHandler) RegisterRoutes(gated *echo.Group) {
	gated.GET("/tours", h.List)
	gated.GET("/tours/:id", h.Get)
	gated.POST("/tours", h.Create)
	gated.PUT("/tours/:id", h.Update)
	gated.PATCH("/tours/:id/status", h.UpdateStatus)
	gated.DELETE("/tours/:id", h.Delete)
}

// tourRequest is the explicit DTO for A3/A4 (never domain.Tour — mass
// assignment: id/slug-uniqueness/timestamps are not directly bindable).
// AvgRating/TotalRatings are only present so BR-005 can detect and reject
// them; they are never written.
type tourRequest struct {
	CategoryID      uuid.UUID             `json:"category_id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Itinerary       *string               `json:"itinerary"`
	Destination     string                `json:"destination"`
	DurationDays    int                   `json:"duration_days"`
	DurationNights  int                   `json:"duration_nights"`
	Price           float64               `json:"price"`
	DiscountPrice   *float64              `json:"discount_price"`
	MaxParticipants int                   `json:"max_participants"`
	ThumbnailURL    *string               `json:"thumbnail_url"`
	Highlights      []string              `json:"highlights"`
	Inclusions      *string               `json:"inclusions"`
	Exclusions      *string               `json:"exclusions"`
	AvgRating       *float64              `json:"avg_rating"`
	TotalRatings    *int                  `json:"total_ratings"`
	Status          *string               `json:"status"`
	Images          []tourImageRequest    `json:"images"`
	Schedules       []tourScheduleRequest `json:"schedules"`
}

func (r tourRequest) toInput() service.TourInput {
	return service.TourInput{
		CategoryID: r.CategoryID, Title: r.Title, Description: r.Description,
		Itinerary: r.Itinerary, Destination: r.Destination, DurationDays: r.DurationDays,
		DurationNights: r.DurationNights, Price: r.Price, DiscountPrice: r.DiscountPrice,
		MaxParticipants: r.MaxParticipants, ThumbnailURL: r.ThumbnailURL, Highlights: r.Highlights,
		Inclusions: r.Inclusions, Exclusions: r.Exclusions, Status: r.Status,
		AvgRating: r.AvgRating, TotalRatings: r.TotalRatings,
	}
}

type tourStatusRequest struct {
	Status string `json:"status"`
}

func (h *TourHandler) List(c echo.Context) error {
	p, err := parseTourListParams(c)
	if err != nil {
		return err
	}
	page, err := h.svc.List(c.Request().Context(), p)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, page)
}

func (h *TourHandler) Get(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	detail, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *TourHandler) Create(c echo.Context) error {
	var req tourRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}

	images := make([]service.TourImageInput, len(req.Images))
	for i, imgReq := range req.Images {
		images[i] = imgReq.toInput()
	}

	schedules := make([]service.TourScheduleInput, len(req.Schedules))
	for i, schReq := range req.Schedules {
		in, err := schReq.toInput()
		if err != nil {
			return err
		}
		schedules[i] = in
	}

	created, err := h.svc.Create(c.Request().Context(), req.toInput(), images, schedules)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *TourHandler) Update(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	var req tourRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	updated, err := h.svc.Update(c.Request().Context(), id, req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *TourHandler) UpdateStatus(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	var req tourStatusRequest
	if err := c.Bind(&req); err != nil || req.Status == "" {
		return apperror.NewBadRequest("status is required")
	}
	updated, err := h.svc.UpdateStatus(c.Request().Context(), id, req.Status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *TourHandler) Delete(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
